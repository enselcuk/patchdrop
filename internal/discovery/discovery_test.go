package discovery

import (
	"net"
	"testing"

	"localshare/internal/protocol"
)

func TestPrefersRoutableWiredAddress(t *testing.T) {
	s := New(0, func() protocol.Announce { return protocol.Announce{ID: "me"} }, func() {})
	s.conn, _ = net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)}) // for replies
	defer s.conn.Close()

	a := protocol.Announce{Magic: protocol.Magic, ID: "peer", Name: "P", Port: 1, Link: "wired", Reply: true}
	ll := protocol.Announce{Magic: protocol.Magic, ID: "peer", Name: "P", Port: 1, Link: "", Reply: true}

	s.handle(ll, &net.UDPAddr{IP: net.ParseIP("169.254.3.4")})
	if got := s.Peers()[0].IP; got != "169.254.3.4" {
		t.Fatalf("first address should be taken, got %s", got)
	}
	s.handle(a, &net.UDPAddr{IP: net.ParseIP("192.168.1.9")})
	if got := s.Peers()[0].IP; got != "192.168.1.9" {
		t.Fatalf("routable wired address should win, got %s", got)
	}
	s.handle(ll, &net.UDPAddr{IP: net.ParseIP("169.254.3.4")})
	if got := s.Peers()[0].IP; got != "192.168.1.9" {
		t.Fatalf("link-local must not replace a fresh routable address, got %s", got)
	}
}
