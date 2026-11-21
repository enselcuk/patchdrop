// Package server is the receiving side: an HTTPS server that accepts offers
// and streams uploaded files to the download directory.
//
// TLS is mutual: both sides present their self-signed device certificate, so
// the receiver knows the sender's fingerprint (used for "auto-accept" of
// trusted devices) and the sender pins the receiver's fingerprint.
package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"patchdrop/internal/identity"
	"patchdrop/internal/protocol"
)

const (
	maxFilesPerOffer = 20000
	sessionIdleLimit = 10 * time.Minute
)

// IncomingOffer is what the UI is asked to accept or decline.
type IncomingOffer struct {
	ID         string              `json:"id"`
	Sender     protocol.Info       `json:"sender"`
	SenderIP   string              `json:"senderIp"`
	Files      []protocol.FileMeta `json:"files"`
	TotalBytes int64               `json:"totalBytes"`
}

type Hooks struct {
	// Info returns this device's info for GET /api/info.
	Info func() protocol.Info
	// DownloadDir returns the directory to save into.
	DownloadDir func() string
	// Decide blocks until the user accepts or declines. ctx ends if the sender disconnects.
	Decide func(ctx context.Context, o IncomingOffer) bool
	// Progress reports bytes received for an accepted session.
	Progress func(sessionID string, received, total int64)
	// Done is called once per accepted session when it finishes, fails or is cancelled.
	Done func(sessionID string, err error, saved []string)
}

type Server struct {
	id    *identity.Identity
	hooks Hooks
	port  int
	srv   *http.Server

	nameMu sync.Mutex // serialises choosing unique file names

	mu       sync.Mutex
	sessions map[string]*session
}

type session struct {
	id     string
	offer  IncomingOffer
	dir    string
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	files    map[string]protocol.FileMeta
	tokens   map[string]string
	pending  int
	received int64
	saved    []string
	lastSeen time.Time
	finished bool
}

func New(id *identity.Identity, hooks Hooks) *Server {
	return &Server{id: id, hooks: hooks, sessions: map[string]*session{}}
}

// Listen starts the server on the first free port in [port, port+10].
func (s *Server) Listen(ctx context.Context, port int) error {
	var ln net.Listener
	var err error
	for p := port; p <= port+10; p++ {
		ln, err = net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%d", p))
		if err == nil {
			s.port = p
			break
		}
	}
	if err != nil {
		return fmt.Errorf("no free port near %d: %w", port, err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", s.handleInfo)
	mux.HandleFunc("POST /api/offer", s.handleOffer)
	mux.HandleFunc("PUT /api/upload", s.handleUpload)
	mux.HandleFunc("POST /api/cancel", s.handleCancel)

	s.srv = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 15 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{s.id.Cert},
			ClientAuth:   tls.RequireAnyClientCert,
			MinVersion:   tls.VersionTLS13,
		},
	}
	go func() { _ = s.srv.ServeTLS(ln, "", "") }()
	go s.reaper(ctx)
	go func() {
		<-ctx.Done()
		sh, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(sh)
	}()
	return nil
}

func (s *Server) Port() int { return s.port }

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.hooks.Info())
}

func clientFingerprint(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}
	return identity.Fingerprint(r.TLS.PeerCertificates[0].Raw)
}

func (s *Server) handleOffer(w http.ResponseWriter, r *http.Request) {
	var offer protocol.Offer
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&offer); err != nil {
		http.Error(w, "bad offer", http.StatusBadRequest)
		return
	}
	if len(offer.Files) == 0 || len(offer.Files) > maxFilesPerOffer {
		http.Error(w, "bad file count", http.StatusBadRequest)
		return
	}
	// The TLS client certificate is the authoritative sender identity.
	offer.Sender.Fingerprint = clientFingerprint(r)

	in := IncomingOffer{ID: randID(), Sender: offer.Sender, Files: offer.Files}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		in.SenderIP = host
	}
	seen := map[string]bool{}
	for _, f := range offer.Files {
		if f.ID == "" || seen[f.ID] || f.Size < 0 {
			http.Error(w, "bad file entry", http.StatusBadRequest)
			return
		}
		if _, err := SanitizeRelPath(f.Name); err != nil {
			http.Error(w, "bad file name: "+f.Name, http.StatusBadRequest)
			return
		}
		seen[f.ID] = true
		in.TotalBytes += f.Size
	}

	if !s.hooks.Decide(r.Context(), in) {
		http.Error(w, "declined", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sess := &session{
		id: in.ID, offer: in, dir: s.hooks.DownloadDir(), ctx: ctx, cancel: cancel,
		files: map[string]protocol.FileMeta{}, tokens: map[string]string{},
		pending: len(in.Files), lastSeen: time.Now(),
	}
	resp := protocol.OfferResponse{Session: in.ID, Tokens: map[string]string{}}
	for _, f := range in.Files {
		tok := randID()
		sess.files[f.ID] = f
		sess.tokens[f.ID] = tok
		resp.Tokens[f.ID] = tok
	}
	s.mu.Lock()
	s.sessions[in.ID] = sess
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sess := s.getSession(q.Get("session"))
	if sess == nil {
		http.Error(w, "unknown session", http.StatusNotFound)
		return
	}
	fileID, token := q.Get("file"), q.Get("token")

	sess.mu.Lock()
	meta, ok := sess.files[fileID]
	want := sess.tokens[fileID]
	if !ok || want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(token)) != 1 {
		sess.mu.Unlock()
		http.Error(w, "bad token", http.StatusForbidden)
		return
	}
	delete(sess.tokens, fileID) // one upload per token
	sess.lastSeen = time.Now()
	sess.mu.Unlock()

	path, err := s.receiveFile(sess, meta, r.Body)
	if err != nil {
		s.finish(sess, err)
		status := http.StatusInternalServerError
		if errors.Is(err, context.Canceled) {
			status = http.StatusGone
		}
		http.Error(w, err.Error(), status)
		return
	}

	sess.mu.Lock()
	sess.saved = append(sess.saved, path)
	sess.pending--
	done := sess.pending == 0
	sess.mu.Unlock()
	if done {
		s.finish(sess, nil)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) receiveFile(sess *session, meta protocol.FileMeta, body io.Reader) (string, error) {
	rel, err := SanitizeRelPath(meta.Name)
	if err != nil {
		return "", err
	}
	final, part, f, err := s.createPart(filepath.Join(sess.dir, rel))
	if err != nil {
		return "", err
	}

	src := &ctxReader{ctx: sess.ctx, r: io.LimitReader(body, meta.Size+1), onRead: func(n int) {
		sess.mu.Lock()
		sess.received += int64(n)
		sess.lastSeen = time.Now()
		rec, total := sess.received, sess.offer.TotalBytes
		sess.mu.Unlock()
		if s.hooks.Progress != nil {
			s.hooks.Progress(sess.id, rec, total)
		}
	}}
	n, err := io.CopyBuffer(f, src, make([]byte, 1<<20))
	cerr := f.Close()
	if err == nil {
		err = cerr
	}
	if err == nil && n != meta.Size {
		err = fmt.Errorf("%s: got %d bytes, expected %d", meta.Name, n, meta.Size)
	}
	if err != nil {
		_ = os.Remove(part)
		return "", err
	}
	if err := os.Rename(part, final); err != nil {
		_ = os.Remove(part)
		return "", err
	}
	return final, nil
}

// createPart picks a non-colliding final name ("x (1).txt") and creates its .part file.
func (s *Server) createPart(want string) (final, part string, f *os.File, err error) {
	s.nameMu.Lock()
	defer s.nameMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
		return "", "", nil, err
	}
	ext := filepath.Ext(want)
	base := strings.TrimSuffix(want, ext)
	for i := 0; i < 10000; i++ {
		final = want
		if i > 0 {
			final = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		part = final + ".part"
		if _, err := os.Lstat(final); err == nil {
			continue
		}
		f, err = os.OpenFile(part, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return final, part, f, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", "", nil, err
		}
	}
	return "", "", nil, fmt.Errorf("could not find a free name for %s", want)
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if sess := s.getSession(r.URL.Query().Get("session")); sess != nil {
		s.finish(sess, context.Canceled)
	}
	w.WriteHeader(http.StatusNoContent)
}

// Cancel aborts an incoming session from the receiver's side.
func (s *Server) Cancel(sessionID string) {
	if sess := s.getSession(sessionID); sess != nil {
		s.finish(sess, context.Canceled)
	}
}

func (s *Server) getSession(id string) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

func (s *Server) finish(sess *session, err error) {
	sess.mu.Lock()
	if sess.finished {
		sess.mu.Unlock()
		return
	}
	sess.finished = true
	saved := append([]string(nil), sess.saved...)
	sess.mu.Unlock()

	sess.cancel()
	s.mu.Lock()
	delete(s.sessions, sess.id)
	s.mu.Unlock()
	if s.hooks.Done != nil {
		s.hooks.Done(sess.id, err, saved)
	}
}

func (s *Server) reaper(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.mu.Lock()
			var stale []*session
			for _, sess := range s.sessions {
				sess.mu.Lock()
				if time.Since(sess.lastSeen) > sessionIdleLimit {
					stale = append(stale, sess)
				}
				sess.mu.Unlock()
			}
			s.mu.Unlock()
			for _, sess := range stale {
				s.finish(sess, errors.New("transfer timed out"))
			}
		}
	}
}

type ctxReader struct {
	ctx    context.Context
	r      io.Reader
	onRead func(int)
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := c.r.Read(p)
	if n > 0 && c.onRead != nil {
		c.onRead(n)
	}
	return n, err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func randID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
