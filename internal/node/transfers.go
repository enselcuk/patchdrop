package node

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"localshare/internal/client"
	"localshare/internal/server"
)

const (
	StatusWaiting   = "waiting" // offer sent, waiting for the receiver to accept
	StatusActive    = "active"
	StatusDone      = "done"
	StatusDeclined  = "declined"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

type Transfer struct {
	ID        string   `json:"id"`
	Direction string   `json:"direction"` // send | receive
	PeerName  string   `json:"peerName"`
	Label     string   `json:"label"`
	Files     int      `json:"files"`
	Total     int64    `json:"total"`
	Done      int64    `json:"done"`
	Speed     float64  `json:"speed"` // bytes/s
	Status    string   `json:"status"`
	Error     string   `json:"error,omitempty"`
	Saved     []string `json:"saved,omitempty"`
	StartedAt int64    `json:"startedAt"` // unix ms
}

type txState struct {
	t        Transfer
	cancel   func()
	lastEmit time.Time
	rateAt   time.Time
	rateB    int64
}

func (n *Node) Transfers() []Transfer {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]Transfer, 0, len(n.tx))
	for _, s := range n.tx {
		out = append(out, s.t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt > out[j].StartedAt })
	return out
}

func (n *Node) addTx(t Transfer, cancel func()) {
	t.StartedAt = time.Now().UnixMilli()
	n.mu.Lock()
	n.tx[t.ID] = &txState{t: t, cancel: cancel, rateAt: time.Now()}
	n.mu.Unlock()
	if n.ev != nil {
		n.ev.TransferUpdated(t)
	}
}

// updateTx mutates a transfer and emits an update (progress is throttled to ~6/s).
func (n *Node) updateTx(id string, fn func(*Transfer)) {
	n.mu.Lock()
	s, ok := n.tx[id]
	if !ok {
		n.mu.Unlock()
		return
	}
	prevStatus := s.t.Status
	fn(&s.t)
	now := time.Now()
	if dt := now.Sub(s.rateAt); dt >= 500*time.Millisecond {
		inst := float64(s.t.Done-s.rateB) / dt.Seconds()
		if s.t.Speed == 0 {
			s.t.Speed = inst
		} else {
			s.t.Speed = 0.6*s.t.Speed + 0.4*inst
		}
		s.rateAt, s.rateB = now, s.t.Done
	}
	emit := s.t.Status != prevStatus || now.Sub(s.lastEmit) >= 150*time.Millisecond
	if emit {
		s.lastEmit = now
	}
	t := s.t
	t.Saved = append([]string(nil), s.t.Saved...)
	n.mu.Unlock()
	if emit && n.ev != nil {
		n.ev.TransferUpdated(t)
	}
}

func finished(status string) bool {
	switch status {
	case StatusDone, StatusDeclined, StatusFailed, StatusCancelled:
		return true
	}
	return false
}

// ClearFinished removes completed / failed transfers from the list.
func (n *Node) ClearFinished() {
	n.mu.Lock()
	for id, s := range n.tx {
		if finished(s.t.Status) {
			delete(n.tx, id)
		}
	}
	n.mu.Unlock()
}

func (n *Node) CancelTransfer(id string) {
	n.mu.Lock()
	s, ok := n.tx[id]
	var cancel func()
	if ok && !finished(s.t.Status) {
		cancel = s.cancel
	}
	n.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func label(rels []string) string {
	if len(rels) == 0 {
		return ""
	}
	top := strings.SplitN(rels[0], "/", 2)[0]
	same := true
	for _, r := range rels[1:] {
		if strings.SplitN(r, "/", 2)[0] != top {
			same = false
			break
		}
	}
	switch {
	case len(rels) == 1:
		return filepath.Base(rels[0])
	case same:
		return top + "/"
	default:
		return fmt.Sprintf("%s + %d more", filepath.Base(rels[0]), len(rels)-1)
	}
}

// ---------- sending ----------

// Send sends files/folders to a known peer in the background and returns the transfer id.
func (n *Node) Send(peerID string, paths []string) (string, error) {
	p, ok := n.peer(peerID)
	if !ok {
		return "", errors.New("unknown device")
	}
	if !p.Online {
		return "", fmt.Errorf("%s is offline", p.Name)
	}
	files, err := client.CollectFiles(paths)
	if err != nil {
		return "", err
	}
	target := client.Target{Host: p.IP, Port: p.Port, Fingerprint: p.Fingerprint}
	// A favorite's pinned fingerprint wins over whatever is announced.
	for _, f := range n.cfg.Get().Favorites {
		if f.ID == p.ID && f.Fingerprint != "" {
			target.Fingerprint = f.Fingerprint
		}
	}

	ctx, cancel := context.WithCancel(n.ctx)
	var total int64
	rels := make([]string, len(files))
	for i, f := range files {
		total += f.Size
		rels[i] = f.Rel
	}
	id := "s-" + fmt.Sprint(time.Now().UnixNano())
	n.addTx(Transfer{ID: id, Direction: "send", PeerName: p.Name, Label: label(rels), Files: len(files),
		Total: total, Status: StatusWaiting}, cancel)

	go func() {
		defer cancel()
		err := n.cli.Send(ctx, target, n.info(), files, func(sent, total int64) {
			n.updateTx(id, func(t *Transfer) { t.Status, t.Done = StatusActive, sent })
		})
		n.updateTx(id, func(t *Transfer) {
			t.Speed = 0
			switch {
			case err == nil:
				t.Status, t.Done = StatusDone, t.Total
			case errors.Is(err, client.ErrDeclined):
				t.Status = StatusDeclined
			case errors.Is(err, context.Canceled):
				t.Status = StatusCancelled
			default:
				t.Status, t.Error = StatusFailed, err.Error()
			}
		})
	}()
	return id, nil
}

// SendTo sends synchronously to an address (CLI use).
func (n *Node) SendTo(ctx context.Context, addr string, paths []string, progress func(sent, total int64)) error {
	host, port, err := ParseAddress(addr)
	if err != nil {
		return err
	}
	files, err := client.CollectFiles(paths)
	if err != nil {
		return err
	}
	return n.cli.Send(ctx, client.Target{Host: host, Port: port}, n.info(), files, progress)
}

// ---------- receiving ----------

func (n *Node) decide(ctx context.Context, o server.IncomingOffer) bool {
	accept := n.opts.AutoAcceptAll
	if !accept {
		for _, f := range n.cfg.Get().Favorites {
			if f.AutoAccept && f.Fingerprint != "" && strings.EqualFold(f.Fingerprint, o.Sender.Fingerprint) {
				accept = true
				break
			}
		}
	}
	if !accept {
		if n.ev == nil {
			return false
		}
		ch := make(chan bool, 1)
		n.mu.Lock()
		n.offers[o.ID] = ch
		n.mu.Unlock()
		n.ev.IncomingOffer(o)

		t := time.NewTimer(offerTimeout)
		select {
		case accept = <-ch:
		case <-ctx.Done():
		case <-t.C:
		}
		t.Stop()
		n.mu.Lock()
		delete(n.offers, o.ID)
		n.mu.Unlock()
		n.ev.OfferClosed(o.ID)
	}
	if !accept {
		return false
	}

	rels := make([]string, len(o.Files))
	for i, f := range o.Files {
		rels[i] = f.Name
	}
	id := o.ID
	n.addTx(Transfer{ID: id, Direction: "receive", PeerName: o.Sender.Name, Label: label(rels), Files: len(o.Files),
		Total: o.TotalBytes, Status: StatusActive}, func() { n.srv.Cancel(id) })
	return true
}

// RespondOffer answers a pending IncomingOffer from the UI.
func (n *Node) RespondOffer(id string, accept bool) {
	n.mu.Lock()
	ch, ok := n.offers[id]
	n.mu.Unlock()
	if ok {
		select {
		case ch <- accept:
		default:
		}
	}
}

func (n *Node) onRecvProgress(id string, received, total int64) {
	n.updateTx(id, func(t *Transfer) { t.Done = received })
}

func (n *Node) onRecvDone(id string, err error, saved []string) {
	n.updateTx(id, func(t *Transfer) {
		t.Speed = 0
		t.Saved = saved
		switch {
		case err == nil:
			t.Status, t.Done = StatusDone, t.Total
		case errors.Is(err, context.Canceled):
			t.Status = StatusCancelled
		default:
			t.Status, t.Error = StatusFailed, err.Error()
		}
	})
}
