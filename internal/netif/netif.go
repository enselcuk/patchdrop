// Package netif enumerates the IPv4 interfaces LocalShare should use for
// discovery. LocalSend-style apps send one multicast packet on whatever
// interface the OS picks (often Wi-Fi or a VPN); we instead send on every real
// interface, which is what makes discovery work on wired-only networks.
package netif

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Kind string

const (
	Wired   Kind = "wired"
	WiFi    Kind = "wifi"
	Unknown Kind = ""
)

type Iface struct {
	Name      string
	Index     int
	IP        net.IP // IPv4
	Net       *net.IPNet
	Broadcast net.IP
	Kind      Kind
}

// List returns usable IPv4 interfaces, wired ones first.
func List() []Iface {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []Iface
	for _, ifc := range ifs {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 || ifc.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		if Skip(ifc.Name) {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipn.IP.To4()
			if ip4 == nil || ip4.IsLoopback() {
				continue
			}
			n := &net.IPNet{IP: ip4.Mask(ipn.Mask), Mask: ipn.Mask}
			out = append(out, Iface{
				Name:      ifc.Name,
				Index:     ifc.Index,
				IP:        ip4,
				Net:       n,
				Broadcast: BroadcastAddr(ip4, ipn.Mask),
				Kind:      kindOf(ifc.Name),
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return rank(out[i].Kind) < rank(out[j].Kind) })
	return out
}

func rank(k Kind) int {
	switch k {
	case Wired:
		return 0
	case Unknown:
		return 1
	default:
		return 2
	}
}

// BroadcastAddr computes the directed broadcast address of ip/mask.
func BroadcastAddr(ip net.IP, mask net.IPMask) net.IP {
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}
	m := mask
	if len(m) == net.IPv6len {
		m = m[12:]
	}
	if len(m) != net.IPv4len {
		return nil
	}
	b := make(net.IP, net.IPv4len)
	for i := range b {
		b[i] = ip4[i] | ^m[i]
	}
	return b
}

// virtual / tunnel / VPN adapters that never reach coworkers' machines.
var skipPrefixes = []string{
	"lo", "utun", "awdl", "llw", "bridge", "gif", "stf", "anpi", "ap", "ipsec", "ppp",
	"docker", "veth", "br-", "virbr", "vmnet", "vboxnet", "tailscale", "zt", "wg", "tun", "tap",
}

var skipContains = []string{
	"vethernet", "virtualbox", "vmware", "hyper-v", "loopback", "tailscale", "zerotier",
	"wireguard", "openvpn", "nordlynx", "pseudo", "wsl", "bluetooth",
}

// Skip reports whether an interface name looks virtual.
func Skip(name string) bool {
	l := strings.ToLower(name)
	for _, p := range skipPrefixes {
		if strings.HasPrefix(l, p) {
			// "ap" prefix is macOS ap1 (Wi-Fi hotspot); don't swallow names like "apple..." on Windows.
			if p == "ap" && runtime.GOOS != "darwin" {
				continue
			}
			return true
		}
	}
	for _, c := range skipContains {
		if strings.Contains(l, c) {
			return true
		}
	}
	return false
}

var (
	darwinOnce  sync.Once
	darwinKinds map[string]Kind
)

func kindOf(name string) Kind {
	switch runtime.GOOS {
	case "darwin":
		darwinOnce.Do(loadDarwinKinds)
		if k, ok := darwinKinds[name]; ok {
			return k
		}
		return Unknown
	case "linux":
		if _, err := os.Stat(filepath.Join("/sys/class/net", name, "wireless")); err == nil {
			return WiFi
		}
		if strings.HasPrefix(name, "en") || strings.HasPrefix(name, "eth") {
			return Wired
		}
		return Unknown
	default: // windows: adapter friendly names
		return KindFromLabel(name)
	}
}

// KindFromLabel classifies a human readable adapter / hardware-port label.
func KindFromLabel(label string) Kind {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "wi-fi"), strings.Contains(l, "wifi"), strings.Contains(l, "wlan"),
		strings.Contains(l, "wireless"), strings.Contains(l, "airport"):
		return WiFi
	case strings.Contains(l, "ethernet"), strings.Contains(l, "lan"), strings.Contains(l, "thunderbolt"),
		strings.Contains(l, "usb 10/100"), strings.Contains(l, "gigabit"):
		return Wired
	}
	return Unknown
}

// loadDarwinKinds parses `networksetup -listallhardwareports`, which maps
// device names (en0, en5…) to hardware port labels (Wi-Fi, Ethernet, …).
func loadDarwinKinds() {
	darwinKinds = map[string]Kind{}
	out, err := exec.Command("/usr/sbin/networksetup", "-listallhardwareports").Output()
	if err != nil {
		return
	}
	var port string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "Hardware Port:"); ok {
			port = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "Device:"); ok && port != "" {
			darwinKinds[strings.TrimSpace(v)] = KindFromLabel(port)
		}
	}
}

// LocalIPs returns this machine's usable IPv4 addresses (for display).
func LocalIPs() []string {
	var ips []string
	for _, i := range List() {
		ips = append(ips, i.IP.String())
	}
	return ips
}
