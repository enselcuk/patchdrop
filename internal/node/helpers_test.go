package node

import (
	"net"
	"strconv"
)

func itoa(i int) string { return strconv.Itoa(i) }

func parseCIDR(s string) (net.IP, *net.IPNet, error) { return net.ParseCIDR(s) }

func mustIP(s string) net.IP { return net.ParseIP(s).To4() }
