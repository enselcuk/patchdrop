// Package client is the sending side: it offers files to a peer and streams
// them over HTTPS, pinning the peer's certificate fingerprint.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"localshare/internal/identity"
	"localshare/internal/protocol"
)

var (
	ErrDeclined            = errors.New("the receiver declined the transfer")
	ErrFingerprintMismatch = errors.New("device identity changed (certificate fingerprint mismatch)")
)

type Client struct {
	id *identity.Identity
}

func New(id *identity.Identity) *Client { return &Client{id: id} }

// Target identifies the receiving device. Fingerprint may be empty for a
// first contact (trust on first use); otherwise it is enforced.
type Target struct {
	Host        string
	Port        int
	Fingerprint string
}

func (t Target) base() string {
	return "https://" + net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
}

func (c *Client) transport(expectFP string, seenFP *string) *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{c.id.Cert},
			MinVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // self-signed; verified by fingerprint below
			VerifyPeerCertificate: func(raw [][]byte, _ [][]*x509.Certificate) error {
				if len(raw) == 0 {
					return errors.New("no server certificate")
				}
				fp := identity.Fingerprint(raw[0])
				if seenFP != nil {
					*seenFP = fp
				}
				if expectFP != "" && !strings.EqualFold(fp, expectFP) {
					return ErrFingerprintMismatch
				}
				return nil
			},
		},
		DialContext:         (&net.Dialer{Timeout: 4 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConnsPerHost: 2,
		WriteBufferSize:     256 << 10,
		ReadBufferSize:      64 << 10,
	}
}

// Info fetches GET /api/info. The returned Fingerprint is the one observed in
// the TLS handshake, not the self-reported one.
func (c *Client) Info(ctx context.Context, t Target) (protocol.Info, error) {
	var seen string
	hc := &http.Client{Transport: c.transport(t.Fingerprint, &seen), Timeout: 5 * time.Second}
	defer hc.CloseIdleConnections()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, t.base()+"/api/info", nil)
	resp, err := hc.Do(req)
	if err != nil {
		return protocol.Info{}, unwrapFP(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return protocol.Info{}, fmt.Errorf("info: %s", resp.Status)
	}
	var info protocol.Info
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&info); err != nil {
		return protocol.Info{}, err
	}
	info.Fingerprint = seen
	return info, nil
}

// LocalFile is a file to send; Rel is the slash-separated name the receiver will use.
type LocalFile struct {
	Path string `json:"path"`
	Rel  string `json:"rel"`
	Size int64  `json:"size"`
}

// CollectFiles expands the given files and folders. A folder "Photos" becomes
// entries "Photos/a.jpg", "Photos/sub/b.jpg", … . Symlinks and OS junk are skipped.
func CollectFiles(paths []string) ([]LocalFile, error) {
	var out []LocalFile
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !st.IsDir() {
			out = append(out, LocalFile{Path: p, Rel: filepath.Base(p), Size: st.Size()})
			continue
		}
		root := filepath.Dir(filepath.Clean(p))
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if junk(d.Name()) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			out = append(out, LocalFile{Path: path, Rel: filepath.ToSlash(rel), Size: info.Size()})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(out) == 0 {
		return nil, errors.New("nothing to send")
	}
	return out, nil
}

func junk(name string) bool {
	switch name {
	case ".DS_Store", "Thumbs.db", "desktop.ini", ".Spotlight-V100", ".Trashes", ".fseventsd":
		return true
	}
	return strings.HasPrefix(name, "._")
}

// Send offers files to the target and uploads them once accepted.
// progress is called with total bytes sent so far.
func (c *Client) Send(ctx context.Context, t Target, self protocol.Info, files []LocalFile, progress func(sent, total int64)) error {
	tr := c.transport(t.Fingerprint, nil)
	hc := &http.Client{Transport: tr}
	defer tr.CloseIdleConnections()

	offer := protocol.Offer{Sender: self}
	var total int64
	for i, f := range files {
		offer.Files = append(offer.Files, protocol.FileMeta{ID: strconv.Itoa(i), Name: f.Rel, Size: f.Size})
		total += f.Size
	}
	body, _ := json.Marshal(offer)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, t.base()+"/api/offer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return unwrapFP(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return ErrDeclined
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		return fmt.Errorf("offer rejected: %s %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	var or protocol.OfferResponse
	err = json.NewDecoder(resp.Body).Decode(&or)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var sent atomic.Int64
	for i, f := range files {
		fid := strconv.Itoa(i)
		if err := c.upload(ctx, hc, t, or, fid, f, func(n int) {
			if progress != nil {
				progress(sent.Add(int64(n)), total)
			}
		}); err != nil {
			if ctx.Err() != nil {
				c.cancelRemote(t, tr, or.Session)
				return context.Canceled
			}
			return err
		}
	}
	if progress != nil {
		progress(total, total)
	}
	return nil
}

func (c *Client) upload(ctx context.Context, hc *http.Client, t Target, or protocol.OfferResponse, fid string, f LocalFile, onRead func(int)) error {
	file, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	q := url.Values{"session": {or.Session}, "file": {fid}, "token": {or.Tokens[fid]}}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, t.base()+"/api/upload?"+q.Encode(),
		&countingReader{r: io.LimitReader(file, f.Size), onRead: onRead})
	req.ContentLength = f.Size
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if resp.StatusCode == http.StatusGone {
			return errors.New("the receiver cancelled the transfer")
		}
		return fmt.Errorf("upload %s: %s %s", f.Rel, resp.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}

func (c *Client) cancelRemote(t Target, tr *http.Transport, session string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, t.base()+"/api/cancel?session="+url.QueryEscape(session), nil)
	if resp, err := (&http.Client{Transport: tr}).Do(req); err == nil {
		resp.Body.Close()
	}
}

type countingReader struct {
	r      io.Reader
	onRead func(int)
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		c.onRead(n)
	}
	return n, err
}

func unwrapFP(err error) error {
	if errors.Is(err, ErrFingerprintMismatch) || strings.Contains(err.Error(), ErrFingerprintMismatch.Error()) {
		return ErrFingerprintMismatch
	}
	return err
}
