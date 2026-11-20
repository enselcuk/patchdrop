package server

import (
	"path/filepath"
	"testing"
)

func TestSanitizeRelPath(t *testing.T) {
	ok := map[string]string{
		"file.txt":         "file.txt",
		"Photos/a.jpg":     filepath.Join("Photos", "a.jpg"),
		"/etc/passwd":      filepath.Join("etc", "passwd"),
		`C:\Windows\x.dll`: filepath.Join("C_", "Windows", "x.dll"),
		"./a/./b":          filepath.Join("a", "b"),
		"what?.txt":        "what_.txt",
		"CON.txt":          "_CON.txt",
		"trailing. ":       "trailing",
		"ünïcødé 文件.pdf":   "ünïcødé 文件.pdf",
	}
	for in, want := range ok {
		got, err := SanitizeRelPath(in)
		if err != nil || got != want {
			t.Errorf("SanitizeRelPath(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "/", "..", "../x", "a/../../x", `..\x`, "./."} {
		if got, err := SanitizeRelPath(bad); err == nil {
			t.Errorf("SanitizeRelPath(%q) = %q; want error", bad, got)
		}
	}
}
