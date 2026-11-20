// Package discovery finds other LocalShare devices on the LAN.
//
// Every few seconds a JSON announce is sent on every usable IPv4 interface to
// that interface's directed broadcast address (e.g. 192.168.1.255) and to a
// multicast group. Directed broadcast is delivered by every switch, so peers on
// wired Ethernet are found even when IGMP snooping drops multicast or the OS
// would otherwise send multicast out of the Wi-Fi / VPN adapter only.
package discovery

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/ipv4"

	"localshare/internal/netif"
	"localshare/internal/protocol"
)

const (
	announceEvery = 3 * time.Second
	peerTTL       = 10 * time.Second
)

type Peer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OS          string    `json:"os"`
	IP          string    `json:"ip"`
	Port        int       `json:"port"`
	Fingerprint string    `json:"fp"`
	Link        string    `json:"link"`
	LastSeen    time.Time `json:"lastSeen"`
	ipSeen      time.Time // last time we heard the peer on IP
}

type Service struct {
	port     int
	self     func() protocol.Announce
	onChange func()

	conn  *net.UDPConn
	pc    *ipv4.PacketConn
	group net.IP

	sendMu sync.Mutex
	joined map[int]bool

	mu        sync.Mutex
	peers     map[string]*Peer
	lastReply map[string]time.Time
}

// New creates a discovery service. self returns this device's current announce
// (its name can change at runtime). onChange is called whenever the peer set changes.
func New(port int, self func() protocol.Announce, onChange func()) *Service {
	return &Service{
		port:      port,
		self:      self,
		onChange:  onChange,
		group:     net.ParseIP(protocol.MulticastGroup).To4(),
		joined:    map[int]bool{},
		peers:     map[string]*Peer{},
		lastReply: map[string]time.Time{},
	}
}

func (s *Service) Start(ctx context.Context) error {
	lc := net.ListenConfig{Control: reuseControl}
	pconn, err := lc.ListenPacket(ctx, "udp4", net.JoinHostPort("0.0.0.0", strconv.Itoa(s.port)))
	if err != nil {
		return err
	}
	s.conn = pconn.(*net.UDPConn)
	s.pc = ipv4.NewPacketConn(s.conn)
	_ = s.pc.SetMulticastLoopback(true)
	_ = s.pc.SetMulticastTTL(1)

	go s.readLoop(ctx)
	go s.announceLoop(ctx)
	go func() {
		<-ctx.Done()
		s.sendAll(true)
		_ = s.conn.Close()
	}()
	return nil
}

// AnnounceNow triggers an immediate announce (e.g. when the user clicks refresh).
func (s *Service) AnnounceNow() { go s.sendAll(false) }

func (s *Service) Peers() []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Peer, 0, len(s.peers))
	for _, p := range s.peers {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Service) announceLoop(ctx context.Context) {
	t := time.NewTicker(announceEvery)
	defer t.Stop()
	s.sendAll(false)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.sendAll(false)
			s.expire()
		}
	}
}

func (s *Service) sendAll(bye bool) {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	ifaces := netif.List()
	seenBcast := map[string]bool{}
	for _, ifc := range ifaces {
		a := s.self()
		a.Link = string(ifc.Kind)
		a.Bye = bye
		msg, _ := json.Marshal(a)

		// 1. directed broadcast on this interface's subnet
		if b := ifc.Broadcast.String(); !seenBcast[b] {
			seenBcast[b] = true
			if _, err := s.conn.WriteToUDP(msg, &net.UDPAddr{IP: ifc.Broadcast, Port: s.port}); err != nil {
				debugf("broadcast on %s (%s): %v", ifc.Name, b, err)
			}
		}

		// 2. multicast out of this specific interface
		ni, err := net.InterfaceByIndex(ifc.Index)
		if err != nil {
			continue
		}
		if !s.joined[ifc.Index] {
			if err := s.pc.JoinGroup(ni, &net.UDPAddr{IP: s.group}); err == nil {
				s.joined[ifc.Index] = true
			}
		}
		if err := s.pc.SetMulticastInterface(ni); err == nil {
			_, _ = s.pc.WriteTo(msg, nil, &net.UDPAddr{IP: s.group, Port: s.port})
		}
	}
	// Limited broadcast as a last resort (reaches the primary interface's segment).
	if len(ifaces) == 0 || bye {
		a := s.self()
		a.Bye = bye
		msg, _ := json.Marshal(a)
		_, _ = s.conn.WriteToUDP(msg, &net.UDPAddr{IP: net.IPv4bcast, Port: s.port})
	}
}

func (s *Service) readLoop(ctx context.Context) {
	buf := make([]byte, 2048)
	for {
		n, src, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		var a protocol.Announce
		if json.Unmarshal(buf[:n], &a) != nil || a.Magic != protocol.Magic || a.ID == "" {
			continue
		}
		self := s.self()
		if a.ID == self.ID {
			continue
		}
		s.handle(a, src)
	}
}

func (s *Service) handle(a protocol.Announce, src *net.UDPAddr) {
	ip := src.IP.To4()
	if ip == nil {
		return
	}
	s.mu.Lock()
	if a.Bye {
		_, existed := s.peers[a.ID]
		delete(s.peers, a.ID)
		s.mu.Unlock()
		if existed {
			s.onChange()
		}
		return
	}
	p, existed := s.peers[a.ID]
	changed := !existed
	if !existed {
		p = &Peer{ID: a.ID}
		s.peers[a.ID] = p
	}
	// A peer with several interfaces is heard on several addresses. Keep the
	// best one (routable + wired beats link-local 169.254.x.x) and only fall
	// back to a worse one when the current address goes quiet.
	sameAddr := p.IP == ip.String()
	useAddr := !existed || sameAddr ||
		addrScore(ip, a.Link) < addrScore(net.ParseIP(p.IP), p.Link) ||
		time.Since(p.ipSeen) > peerTTL/2
	if useAddr {
		if !sameAddr || p.Name != a.Name || p.Port != a.Port || (a.Link != "" && p.Link != a.Link) {
			changed = changed || existed
		}
		p.IP, p.Port = ip.String(), a.Port
		if a.Link != "" {
			p.Link = a.Link
		}
		p.ipSeen = time.Now()
	} else if p.Name != a.Name {
		changed = true
	}
	p.Name, p.OS, p.Fingerprint = a.Name, a.OS, a.Fingerprint
	p.LastSeen = time.Now()

	// Answer new peers directly so both sides appear immediately.
	needReply := !a.Reply && time.Since(s.lastReply[a.ID]) > 2*time.Second
	if needReply {
		s.lastReply[a.ID] = time.Now()
	}
	s.mu.Unlock()

	if needReply {
		r := s.self()
		r.Reply = true
		r.Link = string(linkFor(ip))
		msg, _ := json.Marshal(r)
		_, _ = s.conn.WriteToUDP(msg, src)
	}
	if changed {
		s.onChange()
	}
}

func (s *Service) expire() {
	s.mu.Lock()
	changed := false
	for id, p := range s.peers {
		if time.Since(p.LastSeen) > peerTTL {
			delete(s.peers, id)
			delete(s.lastReply, id)
			changed = true
		}
	}
	s.mu.Unlock()
	if changed {
		s.onChange()
	}
}

// addrScore ranks how good an address is for reaching a peer (lower is better).
func addrScore(ip net.IP, link string) int {
	s := 0
	if ip == nil || ip.IsLinkLocalUnicast() {
		s += 2
	}
	if link != string(netif.Wired) {
		s++
	}
	return s
}

// linkFor returns the kind of the local interface whose subnet contains ip.
func linkFor(ip net.IP) netif.Kind {
	for _, i := range netif.List() {
		if i.Net.Contains(ip) {
			return i.Kind
		}
	}
	return netif.Unknown
}

// Debug enables verbose logging of send errors.
var Debug = false

func debugf(f string, args ...any) {
	if Debug {
		log.Printf("discovery: "+f, args...)
	}
}
