// Package node wires identity, config, discovery, the receive server and the
// send client together. Both the desktop UI (app.go) and the CLI use it.
package node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"localshare/internal/client"
	"localshare/internal/config"
	"localshare/internal/discovery"
	"localshare/internal/identity"
	"localshare/internal/netif"
	"localshare/internal/protocol"
	"localshare/internal/server"
)

const offerTimeout = 2 * time.Minute

// Events receives notifications for the UI. Implementations must not block.
type Events interface {
	PeersChanged(peers []PeerView)
	TransferUpdated(t Transfer)
	IncomingOffer(o server.IncomingOffer)
	OfferClosed(id string)
}

type Options struct {
	Profile       string // separate config dir, for running several instances on one machine
	Port          int    // TCP transfer port (0 = default)
	DiscoveryPort int    // UDP discovery port (0 = default)
	AutoAcceptAll bool   // headless receive mode
}

type PeerView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fp"`
	Link        string `json:"link"`
	Source      string `json:"source"` // lan | manual | scan
	Online      bool   `json:"online"`
	Favorite    bool   `json:"favorite"`
	AutoAccept  bool   `json:"autoAccept"`
}

type SelfInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	OS          string   `json:"os"`
	Port        int      `json:"port"`
	Fingerprint string   `json:"fp"`
	IPs         []string `json:"ips"`
	DownloadDir string   `json:"downloadDir"`
}

type Node struct {
	opts Options
	cfg  *config.Store
	id   *identity.Identity
	disc *discovery.Service
	srv  *server.Server
	cli  *client.Client
	ev   Events

	mu     sync.Mutex
	extra  map[string]*extraPeer // peers found by IP / scan / favorites
	offers map[string]chan bool
	tx     map[string]*txState
	ctx    context.Context
}

type extraPeer struct {
	view   PeerView
	lastOK time.Time
}

func New(opts Options, ev Events) (*Node, error) {
	if opts.Port == 0 {
		opts.Port = protocol.DefaultTransferPort
	}
	if opts.DiscoveryPort == 0 {
		opts.DiscoveryPort = protocol.DiscoveryPort
	}
	dir, err := config.Dir(opts.Profile)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	id, err := identity.LoadOrCreate(dir)
	if err != nil {
		return nil, err
	}
	n := &Node{
		opts: opts, cfg: cfg, id: id, ev: ev, cli: client.New(id),
		extra: map[string]*extraPeer{}, offers: map[string]chan bool{}, tx: map[string]*txState{},
	}
	for _, f := range cfg.Get().Favorites {
		n.extra[f.ID] = &extraPeer{view: PeerView{
			ID: f.ID, Name: f.Name, IP: f.Host, Port: f.Port, Fingerprint: f.Fingerprint, Source: "manual",
		}}
	}
	return n, nil
}

func (n *Node) Start(ctx context.Context) error {
	n.ctx = ctx
	n.srv = server.New(n.id, server.Hooks{
		Info:        n.info,
		DownloadDir: func() string { return n.cfg.Get().DownloadDir },
		Decide:      n.decide,
		Progress:    n.onRecvProgress,
		Done:        n.onRecvDone,
	})
	if err := n.srv.Listen(ctx, n.opts.Port); err != nil {
		return err
	}
	n.disc = discovery.New(n.opts.DiscoveryPort, n.announce, n.peersChanged)
	if err := n.disc.Start(ctx); err != nil {
		return fmt.Errorf("discovery: %w", err)
	}
	go n.healthLoop(ctx)
	return nil
}

func (n *Node) info() protocol.Info {
	s := n.cfg.Get()
	port := 0
	if n.srv != nil {
		port = n.srv.Port()
	}
	return protocol.Info{V: protocol.Version, ID: s.DeviceID, Name: s.DeviceName, OS: runtime.GOOS, Port: port, Fingerprint: n.id.Fingerprint}
}

func (n *Node) announce() protocol.Announce {
	i := n.info()
	return protocol.Announce{Magic: protocol.Magic, V: i.V, ID: i.ID, Name: i.Name, OS: i.OS, Port: i.Port, Fingerprint: i.Fingerprint}
}

func (n *Node) Self() SelfInfo {
	i := n.info()
	return SelfInfo{ID: i.ID, Name: i.Name, OS: i.OS, Port: i.Port, Fingerprint: i.Fingerprint,
		IPs: netif.LocalIPs(), DownloadDir: n.cfg.Get().DownloadDir}
}

// ---------- peers ----------

func (n *Node) Peers() []PeerView {
	favs := map[string]config.Favorite{}
	for _, f := range n.cfg.Get().Favorites {
		favs[f.ID] = f
	}
	byID := map[string]PeerView{}

	n.mu.Lock()
	for id, e := range n.extra {
		v := e.view
		v.Online = time.Since(e.lastOK) < 15*time.Second
		byID[id] = v
	}
	n.mu.Unlock()

	if n.disc != nil {
		for _, p := range n.disc.Peers() {
			byID[p.ID] = PeerView{ID: p.ID, Name: p.Name, OS: p.OS, IP: p.IP, Port: p.Port,
				Fingerprint: p.Fingerprint, Link: p.Link, Source: "lan", Online: true}
		}
	}
	out := make([]PeerView, 0, len(byID))
	for id, v := range byID {
		if f, ok := favs[id]; ok {
			v.Favorite, v.AutoAccept = true, f.AutoAccept
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Online != out[j].Online {
			return out[i].Online
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func (n *Node) peer(id string) (PeerView, bool) {
	for _, p := range n.Peers() {
		if p.ID == id {
			return p, true
		}
	}
	return PeerView{}, false
}

func (n *Node) peersChanged() {
	if n.ev != nil {
		n.ev.PeersChanged(n.Peers())
	}
}

// Refresh re-announces and re-checks manually added devices.
func (n *Node) Refresh() {
	if n.disc != nil {
		n.disc.AnnounceNow()
	}
	go n.checkExtras(n.ctx)
}

// ParseAddress accepts "host", "host:port" or "[v6]:port".
func ParseAddress(addr string) (string, int, error) {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimPrefix(strings.TrimPrefix(addr, "https://"), "http://")
	addr = strings.TrimSuffix(addr, "/")
	if addr == "" {
		return "", 0, errors.New("enter an IP address")
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, protocol.DefaultTransferPort, nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port %q", portStr)
	}
	return host, port, nil
}

// AddByAddress contacts a device directly (no discovery needed) and saves it as a favorite.
func (n *Node) AddByAddress(ctx context.Context, addr string) (PeerView, error) {
	host, port, err := ParseAddress(addr)
	if err != nil {
		return PeerView{}, err
	}
	info, err := n.cli.Info(ctx, client.Target{Host: host, Port: port})
	if err != nil {
		return PeerView{}, fmt.Errorf("could not reach %s:%d – is LocalShare running there and allowed through the firewall? (%v)", host, port, err)
	}
	if info.ID == n.info().ID {
		return PeerView{}, errors.New("that address is this device")
	}
	v := PeerView{ID: info.ID, Name: info.Name, OS: info.OS, IP: host, Port: port, Fingerprint: info.Fingerprint, Source: "manual"}
	n.mu.Lock()
	n.extra[info.ID] = &extraPeer{view: v, lastOK: time.Now()}
	n.mu.Unlock()
	if err := n.upsertFavorite(v, nil); err != nil {
		return PeerView{}, err
	}
	n.peersChanged()
	v.Online, v.Favorite = true, true
	return v, nil
}

// SetFavorite stars / un-stars a device. autoAccept lets it send without a prompt.
func (n *Node) SetFavorite(peerID string, fav, autoAccept bool) error {
	p, ok := n.peer(peerID)
	if !ok {
		return errors.New("unknown device")
	}
	if fav {
		if err := n.upsertFavorite(p, &autoAccept); err != nil {
			return err
		}
	} else {
		err := n.cfg.Update(func(s *config.Settings) {
			out := s.Favorites[:0]
			for _, f := range s.Favorites {
				if f.ID != peerID {
					out = append(out, f)
				}
			}
			s.Favorites = out
		})
		if err != nil {
			return err
		}
		n.mu.Lock()
		if e, ok := n.extra[peerID]; ok && e.view.Source == "manual" {
			delete(n.extra, peerID)
		}
		n.mu.Unlock()
	}
	n.peersChanged()
	return nil
}

func (n *Node) upsertFavorite(p PeerView, autoAccept *bool) error {
	return n.cfg.Update(func(s *config.Settings) {
		for i := range s.Favorites {
			if s.Favorites[i].ID == p.ID {
				f := &s.Favorites[i]
				f.Name, f.Host, f.Port = p.Name, p.IP, p.Port
				if f.Fingerprint == "" {
					f.Fingerprint = p.Fingerprint
				}
				if autoAccept != nil {
					f.AutoAccept = *autoAccept
				}
				return
			}
		}
		f := config.Favorite{ID: p.ID, Name: p.Name, Host: p.IP, Port: p.Port, Fingerprint: p.Fingerprint}
		if autoAccept != nil {
			f.AutoAccept = *autoAccept
		}
		s.Favorites = append(s.Favorites, f)
	})
}

// healthLoop pings manually-added / scanned devices so their online state stays current.
func (n *Node) healthLoop(ctx context.Context) {
	n.checkExtras(ctx)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n.checkExtras(ctx)
		}
	}
}

func (n *Node) checkExtras(ctx context.Context) {
	if ctx == nil {
		return
	}
	lan := map[string]bool{}
	if n.disc != nil {
		for _, p := range n.disc.Peers() {
			lan[p.ID] = true
		}
	}
	n.mu.Lock()
	var todo []PeerView
	for id, e := range n.extra {
		if !lan[id] {
			todo = append(todo, e.view)
		}
	}
	n.mu.Unlock()

	var wg sync.WaitGroup
	changed := false
	var cmu sync.Mutex
	for _, v := range todo {
		wg.Add(1)
		go func(v PeerView) {
			defer wg.Done()
			c, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			info, err := n.cli.Info(c, client.Target{Host: v.IP, Port: v.Port, Fingerprint: v.Fingerprint})
			n.mu.Lock()
			defer n.mu.Unlock()
			e, ok := n.extra[v.ID]
			if !ok {
				return
			}
			wasOnline := time.Since(e.lastOK) < 15*time.Second
			renamed := false
			if err == nil && info.ID == v.ID {
				e.lastOK = time.Now()
				renamed = e.view.Name != info.Name
				e.view.Name, e.view.OS = info.Name, info.OS
			} else if e.view.Source == "scan" && time.Since(e.lastOK) > time.Minute {
				delete(n.extra, v.ID)
			}
			if renamed || (time.Since(e.lastOK) < 15*time.Second) != wasOnline {
				cmu.Lock()
				changed = true
				cmu.Unlock()
			}
		}(v)
	}
	wg.Wait()
	if changed {
		n.peersChanged()
	}
}

// Scan probes every address of each local subnet (max /22, otherwise the
// surrounding /24) for a LocalShare server. Use it when broadcast is blocked.
func (n *Node) Scan(ctx context.Context) (int, error) {
	self := n.info()
	var hosts []string
	seen := map[string]bool{}
	for _, ifc := range netif.List() {
		for _, h := range subnetHosts(ifc.IP, ifc.Net) {
			if !seen[h] && h != ifc.IP.String() {
				seen[h] = true
				hosts = append(hosts, h)
			}
		}
	}
	if len(hosts) == 0 {
		return 0, errors.New("no network interface with an IPv4 address")
	}

	ports := []int{protocol.DefaultTransferPort, protocol.DefaultTransferPort + 1, protocol.DefaultTransferPort + 2}
	type hit struct {
		host string
		port int
	}
	jobs := make(chan hit)
	var hits []hit
	var hmu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < 128; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d := net.Dialer{Timeout: 400 * time.Millisecond}
			for j := range jobs {
				c, err := d.DialContext(ctx, "tcp4", net.JoinHostPort(j.host, strconv.Itoa(j.port)))
				if err == nil {
					c.Close()
					hmu.Lock()
					hits = append(hits, j)
					hmu.Unlock()
				}
			}
		}()
	}
feed:
	for _, h := range hosts {
		for _, p := range ports {
			select {
			case jobs <- hit{h, p}:
			case <-ctx.Done():
				break feed
			}
		}
	}
	close(jobs)
	wg.Wait()

	found := 0
	for _, h := range hits {
		c, cancel := context.WithTimeout(ctx, 2*time.Second)
		info, err := n.cli.Info(c, client.Target{Host: h.host, Port: h.port})
		cancel()
		if err != nil || info.ID == "" || info.ID == self.ID {
			continue
		}
		found++
		n.mu.Lock()
		if e, ok := n.extra[info.ID]; ok {
			e.lastOK = time.Now()
			e.view.IP, e.view.Port = h.host, h.port
		} else {
			n.extra[info.ID] = &extraPeer{lastOK: time.Now(), view: PeerView{
				ID: info.ID, Name: info.Name, OS: info.OS, IP: h.host, Port: h.port, Fingerprint: info.Fingerprint, Source: "scan",
			}}
		}
		n.mu.Unlock()
	}
	n.peersChanged()
	return found, ctx.Err()
}

func subnetHosts(ip net.IP, ipn *net.IPNet) []string {
	ones, bits := ipn.Mask.Size()
	if bits != 32 {
		return nil
	}
	base := ipn.IP.To4()
	if ones < 22 { // too big – scan only the /24 around us
		ones = 24
		base = ip.To4().Mask(net.CIDRMask(24, 32))
	}
	size := uint32(1) << (32 - ones)
	start := uint32(base[0])<<24 | uint32(base[1])<<16 | uint32(base[2])<<8 | uint32(base[3])
	var out []string
	for i := uint32(1); i+1 < size; i++ {
		v := start + i
		out = append(out, net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v)).String())
	}
	return out
}

// ---------- settings ----------

func (n *Node) SetDeviceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return errors.New("name must be 1–64 characters")
	}
	if err := n.cfg.Update(func(s *config.Settings) { s.DeviceName = name }); err != nil {
		return err
	}
	if n.disc != nil {
		n.disc.AnnounceNow()
	}
	return nil
}

func (n *Node) SetDownloadDir(dir string) error {
	if dir == "" {
		return errors.New("empty folder")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return n.cfg.Update(func(s *config.Settings) { s.DownloadDir = dir })
}
