package client_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"localshare/internal/client"
	"localshare/internal/identity"
	"localshare/internal/protocol"
	"localshare/internal/server"
)

type harness struct {
	recvDir  string
	srv      *server.Server
	sender   *client.Client
	senderID *identity.Identity
	recvID   *identity.Identity
	accept   bool
	mu       sync.Mutex
	offers   []server.IncomingOffer
	done     chan error
}

func newHarness(t *testing.T, accept bool) *harness {
	t.Helper()
	h := &harness{recvDir: t.TempDir(), accept: accept, done: make(chan error, 1)}
	var err error
	if h.recvID, err = identity.LoadOrCreate(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if h.senderID, err = identity.LoadOrCreate(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	h.sender = client.New(h.senderID)
	h.srv = server.New(h.recvID, server.Hooks{
		Info:        func() protocol.Info { return protocol.Info{V: 1, ID: "recv", Name: "Receiver"} },
		DownloadDir: func() string { return h.recvDir },
		Decide: func(ctx context.Context, o server.IncomingOffer) bool {
			h.mu.Lock()
			h.offers = append(h.offers, o)
			h.mu.Unlock()
			return h.accept
		},
		Done: func(id string, err error, saved []string) { h.done <- err },
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := h.srv.Listen(ctx, 48921); err != nil {
		t.Fatal(err)
	}
	return h
}

func (h *harness) target() client.Target {
	return client.Target{Host: "127.0.0.1", Port: h.srv.Port(), Fingerprint: h.recvID.Fingerprint}
}

func writeRandom(t *testing.T, path string, size int) []byte {
	t.Helper()
	b := make([]byte, size)
	_, _ = rand.Read(b)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSendFilesAndFolder(t *testing.T) {
	h := newHarness(t, true)
	src := t.TempDir()
	big := writeRandom(t, filepath.Join(src, "big.bin"), 20<<20+123)
	a := writeRandom(t, filepath.Join(src, "Project", "a.txt"), 1000)
	b := writeRandom(t, filepath.Join(src, "Project", "sub", "b.txt"), 0)
	writeRandom(t, filepath.Join(src, "Project", ".DS_Store"), 10)

	files, err := client.CollectFiles([]string{filepath.Join(src, "big.bin"), filepath.Join(src, "Project")})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("want 3 files (junk skipped), got %d: %+v", len(files), files)
	}

	var last int64
	err = h.sender.Send(context.Background(), h.target(), protocol.Info{ID: "send", Name: "Sender"}, files,
		func(sent, total int64) { last = sent })
	if err != nil {
		t.Fatal(err)
	}
	if err := <-h.done; err != nil {
		t.Fatal(err)
	}
	if want := int64(len(big) + len(a) + len(b)); last != want {
		t.Errorf("progress ended at %d, want %d", last, want)
	}
	check := func(rel string, want []byte) {
		got, err := os.ReadFile(filepath.Join(h.recvDir, rel))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if sha256.Sum256(got) != sha256.Sum256(want) {
			t.Errorf("%s: content mismatch", rel)
		}
	}
	check("big.bin", big)
	check(filepath.Join("Project", "a.txt"), a)
	check(filepath.Join("Project", "sub", "b.txt"), b)

	// The receiver learned the sender's real fingerprint from mutual TLS.
	if fp := h.offers[0].Sender.Fingerprint; fp != h.senderID.Fingerprint {
		t.Errorf("sender fingerprint %q, want %q", fp, h.senderID.Fingerprint)
	}

	// Sending the same file again must not overwrite: "big (1).bin".
	one, _ := client.CollectFiles([]string{filepath.Join(src, "big.bin")})
	if err := h.sender.Send(context.Background(), h.target(), protocol.Info{}, one, nil); err != nil {
		t.Fatal(err)
	}
	<-h.done
	if _, err := os.Stat(filepath.Join(h.recvDir, "big (1).bin")); err != nil {
		t.Errorf("expected renamed duplicate: %v", err)
	}
	assertNoPartFiles(t, h.recvDir)
}

func TestDeclined(t *testing.T) {
	h := newHarness(t, false)
	src := t.TempDir()
	writeRandom(t, filepath.Join(src, "x.txt"), 10)
	files, _ := client.CollectFiles([]string{filepath.Join(src, "x.txt")})
	err := h.sender.Send(context.Background(), h.target(), protocol.Info{}, files, nil)
	if !errors.Is(err, client.ErrDeclined) {
		t.Fatalf("want ErrDeclined, got %v", err)
	}
}

func TestFingerprintPinning(t *testing.T) {
	h := newHarness(t, true)
	tg := h.target()
	tg.Fingerprint = strings.Repeat("0", 64)
	if _, err := h.sender.Info(context.Background(), tg); !errors.Is(err, client.ErrFingerprintMismatch) {
		t.Fatalf("want fingerprint mismatch, got %v", err)
	}
	tg.Fingerprint = ""
	info, err := h.sender.Info(context.Background(), tg)
	if err != nil || info.Fingerprint != h.recvID.Fingerprint {
		t.Fatalf("TOFU info: %+v %v", info, err)
	}
}

func TestCancelMidTransfer(t *testing.T) {
	h := newHarness(t, true)
	src := t.TempDir()
	writeRandom(t, filepath.Join(src, "huge.bin"), 200<<20)
	files, _ := client.CollectFiles([]string{filepath.Join(src, "huge.bin")})

	ctx, cancel := context.WithCancel(context.Background())
	var once sync.Once
	err := h.sender.Send(ctx, h.target(), protocol.Info{}, files, func(sent, total int64) {
		if sent > 5<<20 {
			once.Do(cancel)
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancelled, got %v", err)
	}
	select {
	case <-h.done:
	case <-time.After(5 * time.Second):
		t.Fatal("receiver never finished the cancelled session")
	}
	time.Sleep(100 * time.Millisecond)
	entries, _ := os.ReadDir(h.recvDir)
	if len(entries) != 0 {
		t.Errorf("expected no leftover files, got %v", entries)
	}
}

func assertNoPartFiles(t *testing.T, dir string) {
	t.Helper()
	_ = filepath.Walk(dir, func(p string, _ os.FileInfo, _ error) error {
		if strings.HasSuffix(p, ".part") {
			t.Errorf("leftover part file %s", p)
		}
		return nil
	})
}
