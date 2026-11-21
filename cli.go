package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"patchdrop/internal/discovery"
	"patchdrop/internal/netif"
	"patchdrop/internal/node"
	"patchdrop/internal/server"
)

func isCLICommand(arg string) bool {
	switch arg {
	case "peers", "send", "receive", "interfaces", "help", "-h", "--help":
		return true
	}
	return false
}

const usage = `PatchDrop – share files over the local network

Usage:
  patchdrop                         start the desktop app
  patchdrop peers                   list devices found on the network (8 s)
  patchdrop send <ip[:port]> <file|folder>...
                                     send to a device by address
  patchdrop receive                 headless receiver; accepts everything
  patchdrop interfaces              show the interfaces used for discovery

Environment:
  PATCHDROP_PROFILE   use a separate settings folder (run several instances)
  PATCHDROP_PORT      transfer port (default 47821)
`

type cliEvents struct{}

func (cliEvents) PeersChanged([]node.PeerView) {}
func (cliEvents) TransferUpdated(t node.Transfer) {
	if t.Direction == "receive" && t.Status != node.StatusActive {
		fmt.Printf("%s: %s from %s (%s)%s\n", t.Status, t.Label, t.PeerName, humanBytes(t.Total), errSuffix(t.Error))
	}
}
func (cliEvents) IncomingOffer(server.IncomingOffer) {}
func (cliEvents) OfferClosed(string)                 {}

func errSuffix(e string) string {
	if e == "" {
		return ""
	}
	return " – " + e
}

func runCLI(opts node.Options, args []string) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0

	case "interfaces":
		for _, i := range netif.List() {
			kind := string(i.Kind)
			if kind == "" {
				kind = "unknown"
			}
			fmt.Printf("%-10s %-16s %-18s broadcast %-16s %s\n", i.Name, i.IP, i.Net, i.Broadcast, kind)
		}
		return 0

	case "peers":
		discovery.Debug = os.Getenv("PATCHDROP_DEBUG") != ""
		n, err := node.New(opts, nil)
		if err == nil {
			err = n.Start(ctx)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("Searching for 8 seconds…")
		select {
		case <-time.After(8 * time.Second):
		case <-ctx.Done():
		}
		peers := n.Peers()
		if len(peers) == 0 {
			fmt.Println("No devices found. Check the firewall on both machines, or use: patchdrop send <ip> <file>")
		}
		for _, p := range peers {
			state := "offline"
			if p.Online {
				state = "online"
			}
			fmt.Printf("%-24s %-8s %s:%d  %-6s %-6s %s\n", p.Name, p.OS, p.IP, p.Port, p.Link, p.Source, state)
		}
		return 0

	case "send":
		if len(args) < 3 {
			fmt.Fprint(os.Stderr, usage)
			return 2
		}
		n, err := node.New(opts, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Printf("Waiting for %s to accept…\n", args[1])
		start := time.Now()
		var lastPrint time.Time
		err = n.SendTo(ctx, args[1], args[2:], func(sent, total int64) {
			if time.Since(lastPrint) < 200*time.Millisecond && sent != total {
				return
			}
			lastPrint = time.Now()
			pct := 100.0
			if total > 0 {
				pct = float64(sent) * 100 / float64(total)
			}
			speed := float64(sent) / time.Since(start).Seconds()
			fmt.Printf("\r%5.1f%%  %s / %s  %s/s   ", pct, humanBytes(sent), humanBytes(total), humanBytes(int64(speed)))
		})
		fmt.Println()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Printf("Done in %s\n", time.Since(start).Round(time.Millisecond))
		return 0

	case "receive":
		opts.AutoAcceptAll = true
		n, err := node.New(opts, cliEvents{})
		if err == nil {
			err = n.Start(ctx)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		s := n.Self()
		fmt.Printf("Receiving as %q on %s port %d\nSaving to %s\nPress Ctrl+C to stop.\n",
			s.Name, strings.Join(s.IPs, ", "), s.Port, s.DownloadDir)
		<-ctx.Done()
		return 0
	}
	return 2
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
