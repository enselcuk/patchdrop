package server

import (
	"errors"
	"path/filepath"
	"strings"
)

var errBadName = errors.New("invalid file name")

var reservedWindows = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// SanitizeRelPath turns a sender-supplied relative path ("folder/file.txt")
// into a safe OS path that can never escape the download directory and is
// valid on both macOS and Windows.
func SanitizeRelPath(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	var parts []string
	for _, p := range strings.Split(name, "/") {
		p = strings.TrimSpace(p)
		if p == "" || p == "." {
			continue
		}
		if p == ".." {
			return "", errBadName
		}
		p = strings.Map(func(r rune) rune {
			if r < 0x20 || strings.ContainsRune(`<>:"|?*`, r) {
				return '_'
			}
			return r
		}, p)
		p = strings.TrimRight(p, ". ") // Windows drops trailing dots/spaces
		if p == "" {
			p = "_"
		}
		stem := strings.ToUpper(strings.SplitN(p, ".", 2)[0])
		if reservedWindows[stem] {
			p = "_" + p
		}
		if len(p) > 240 {
			p = strings.ToValidUTF8(p[:240], "")
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 || len(parts) > 64 {
		return "", errBadName
	}
	return filepath.Join(parts...), nil
}
