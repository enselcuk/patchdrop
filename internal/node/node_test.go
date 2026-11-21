package node

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"patchdrop/internal/server"
)

type recorder struct {
	mu    sync.Mutex
	tx    map[string]Transfer
	n     *Node
	allow bool
}

func (r *recorder) PeersChanged([]PeerView) {}
func (r *recorder) TransferUpdated(t Transfer) {
	r.mu.Lock()
	r.tx[t.ID] = t
	r.mu.Unlock()
}
func (r *recorder) IncomingOffer(o server.IncomingOffer) { go r.n.RespondOffer(o.ID, r.allow) }
func (r *recorder) OfferClosed(string)                   {}

func (r *recorder) status(id string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tx[id].Status
}

func startNode(t *testing.T, ctx context.Context, profile string, port int, allow bool) (*Node, *recorder) {
	t.Helper()
	rec := &recorder{tx: map[string]Transfer{}, allow: allow}
	n, err := New(Options{Profile: profile, Port: port, DiscoveryPort: 47990}, rec)
	if err != nil {
		t.Fatal(err)
	}
	rec.n = n
	if err := n.SetDownloadDir(filepath.Join(t.TempDir(), profile)); err != nil {
		t.Fatal(err)
	}
	if err := n.Start(ctx); err != nil {
		t.Fatal(err)
	}
	return n, rec
}

func waitFor(t *testing.T, what string, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// Two nodes on one machine discover each other via UDP broadcast/multicast
// and transfer a file through the full stack (discovery → offer → prompt → upload).
func TestTwoNodesDiscoverAndSend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("AppData", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, _ := startNode(t, ctx, "a", 48831, true)
	b, recB := startNode(t, ctx, "b", 48841, true)

	find := func(n *Node, id string) (PeerView, bool) {
		for _, p := range n.Peers() {
			if p.ID == id && p.Online {
				return p, true
			}
		}
		return PeerView{}, false
	}
	waitFor(t, "a to see b", 8*time.Second, func() bool { _, ok := find(a, b.Self().ID); return ok })
	waitFor(t, "b to see a", 8*time.Second, func() bool { _, ok := find(b, a.Self().ID); return ok })
	pb, _ := find(a, b.Self().ID)
	if pb.Source != "lan" || pb.Fingerprint != b.Self().Fingerprint {
		t.Fatalf("unexpected peer view %+v", pb)
	}

	src := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(src, []byte("hello over the wire"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := a.Send(b.Self().ID, []string{src})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "receive done", 10*time.Second, func() bool {
		for _, tr := range b.Transfers() {
			if tr.Direction == "receive" && tr.Status == StatusDone {
				return true
			}
		}
		return false
	})
	waitFor(t, "send done", 5*time.Second, func() bool {
		for _, tr := range a.Transfers() {
			if tr.ID == id {
				return tr.Status == StatusDone
			}
		}
		return false
	})
	got, err := os.ReadFile(filepath.Join(b.Self().DownloadDir, "hello.txt"))
	if err != nil || string(got) != "hello over the wire" {
		t.Fatalf("received %q, %v", got, err)
	}
	_ = recB

	// Manual add-by-IP path (what users do when broadcast is blocked).
	pv, err := a.AddByAddress(ctx, "127.0.0.1:"+itoa(b.Self().Port))
	if err != nil || pv.ID != b.Self().ID || !pv.Favorite {
		t.Fatalf("AddByAddress: %+v %v", pv, err)
	}
}

func TestParseAddress(t *testing.T) {
	cases := map[string]struct {
		host string
		port int
	}{
		"192.168.1.5":               {"192.168.1.5", 47821},
		"192.168.1.5:5000":          {"192.168.1.5", 5000},
		" https://10.0.0.2:47822/ ": {"10.0.0.2", 47822},
		"mac-mini.local":            {"mac-mini.local", 47821},
	}
	for in, want := range cases {
		h, p, err := ParseAddress(in)
		if err != nil || h != want.host || p != want.port {
			t.Errorf("ParseAddress(%q) = %q,%d,%v", in, h, p, err)
		}
	}
	if _, _, err := ParseAddress("1.2.3.4:99999"); err == nil {
		t.Error("expected invalid port error")
	}
}

func TestSubnetHosts(t *testing.T) {
	_, n24, _ := parseCIDR("192.168.1.0/24")
	if got := subnetHosts(n24.IP, n24); len(got) != 254 || got[0] != "192.168.1.1" || got[253] != "192.168.1.254" {
		t.Errorf("/24: %d hosts, first %v", len(got), got[:1])
	}
	_, n16, _ := parseCIDR("10.1.0.0/16")
	ip := mustIP("10.1.7.9")
	if got := subnetHosts(ip, n16); len(got) != 254 || got[0] != "10.1.7.1" {
		t.Errorf("/16 should collapse to /24 around us, got %d hosts first %v", len(got), got[:1])
	}
}

// The UI declining an offer must reach the sender as a decline and save nothing.
func TestDeclineFromUI(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, _ := startNode(t, ctx, "da", 48851, true)
	b, _ := startNode(t, ctx, "db", 48861, false) // b's UI declines

	src := filepath.Join(t.TempDir(), "nope.txt")
	_ = os.WriteFile(src, []byte("x"), 0o644)
	err := a.SendTo(ctx, "127.0.0.1:"+itoa(b.Self().Port), []string{src}, nil)
	if err == nil || err.Error() != "the receiver declined the transfer" {
		t.Fatalf("want decline, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(b.Self().DownloadDir, "nope.txt")); err == nil {
		t.Fatal("declined file was saved")
	}
}
