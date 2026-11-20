package netif

import (
	"net"
	"testing"
)

func TestBroadcastAddr(t *testing.T) {
	cases := []struct{ ip, mask, want string }{
		{"192.168.1.23", "255.255.255.0", "192.168.1.255"},
		{"10.0.5.9", "255.255.0.0", "10.0.255.255"},
		{"172.16.3.130", "255.255.255.128", "172.16.3.255"},
		{"169.254.10.20", "255.255.0.0", "169.254.255.255"},
	}
	for _, c := range cases {
		mask := net.IPMask(net.ParseIP(c.mask).To4())
		got := BroadcastAddr(net.ParseIP(c.ip), mask)
		if got.String() != c.want {
			t.Errorf("BroadcastAddr(%s,%s)=%s want %s", c.ip, c.mask, got, c.want)
		}
	}
}

func TestSkip(t *testing.T) {
	skip := []string{"lo0", "utun3", "awdl0", "llw0", "bridge100", "docker0", "vEthernet (WSL)", "VirtualBox Host-Only Network", "Tailscale"}
	keep := []string{"en0", "en7", "Ethernet", "Ethernet 2", "Wi-Fi", "eth0"}
	for _, n := range skip {
		if !Skip(n) {
			t.Errorf("expected %q to be skipped", n)
		}
	}
	for _, n := range keep {
		if Skip(n) {
			t.Errorf("expected %q to be kept", n)
		}
	}
}

func TestKindFromLabel(t *testing.T) {
	cases := map[string]Kind{
		"Wi-Fi":                WiFi,
		"WLAN":                 WiFi,
		"Ethernet":             Wired,
		"Ethernet 2":           Wired,
		"Thunderbolt Ethernet": Wired,
		"USB 10/100/1000 LAN":  Wired,
		"Thunderbolt Bridge":   Wired,
		"iPhone USB":           Unknown,
	}
	for in, want := range cases {
		if got := KindFromLabel(in); got != want {
			t.Errorf("KindFromLabel(%q)=%q want %q", in, got, want)
		}
	}
}
