// Package config stores user settings as JSON in the OS config directory.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Favorite is a device remembered by the user (added by IP or starred).
type Favorite struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fp"`
	AutoAccept  bool   `json:"autoAccept"`
}

type Settings struct {
	DeviceID    string     `json:"deviceId"`
	DeviceName  string     `json:"deviceName"`
	DownloadDir string     `json:"downloadDir"`
	Favorites   []Favorite `json:"favorites"`
}

type Store struct {
	mu   sync.Mutex
	dir  string
	path string
	s    Settings
}

// Dir returns the config directory. A non-empty profile (PATCHDROP_PROFILE)
// gives a separate directory, which allows running several instances on one machine.
func Dir(profile string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	name := "PatchDrop"
	if profile != "" {
		name += "-" + profile
	}
	return filepath.Join(base, name), nil
}

func Load(dir string) (*Store, error) {
	st := &Store{dir: dir, path: filepath.Join(dir, "config.json")}
	data, err := os.ReadFile(st.path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &st.s); err != nil {
			return nil, err
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return nil, err
	}

	changed := false
	if st.s.DeviceID == "" {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		st.s.DeviceID = hex.EncodeToString(b)
		changed = true
	}
	if st.s.DeviceName == "" {
		st.s.DeviceName = defaultName()
		changed = true
	}
	if st.s.DownloadDir == "" {
		home, _ := os.UserHomeDir()
		st.s.DownloadDir = filepath.Join(home, "Downloads", "PatchDrop")
		changed = true
	}
	if changed {
		if err := st.saveLocked(); err != nil {
			return nil, err
		}
	}
	return st, nil
}

func defaultName() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "PatchDrop device"
	}
	h = strings.TrimSuffix(h, ".local")
	h = strings.TrimSuffix(h, ".lan")
	return h
}

func (st *Store) Get() Settings {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.s
	s.Favorites = append([]Favorite(nil), st.s.Favorites...)
	return s
}

// Update applies fn to the settings and persists them.
func (st *Store) Update(fn func(*Settings)) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	fn(&st.s)
	return st.saveLocked()
}

func (st *Store) saveLocked() error {
	if err := os.MkdirAll(st.dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st.s, "", "  ")
	if err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}
