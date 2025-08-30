// filename: internal/ci/covercheck/run_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	os.WriteFile(profile, []byte("mode: set\nfile.go:1.1,1.2 1 1\n"), 0o644)
	var out, errBuf bytes.Buffer
	code := run([]string{"covercheck", profile}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Total coverage: 100.0%") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    string
	}{
		{"noargs", "", "usage"},
		{"empty", "mode: set", "no statements"},
		{"invalid", "mode: set\nfile.go:1.1,1.2 x y\n", "invalid line"},
		{"below", "mode: set\nfile.go:1.1,1.2 1 0\n", "Coverage 0.0% is below"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			var args []string
			if tc.profile != "" {
				path := filepath.Join(dir, "cov.out")
				os.WriteFile(path, []byte(tc.profile), 0o644)
				args = []string{"covercheck", path}
			} else {
				args = []string{"covercheck"}
			}
			var out, errBuf bytes.Buffer
			code := run(args, &out, &errBuf)
			if code == 0 {
				t.Fatalf("expected non-zero exit")
			}
			if !strings.Contains(out.String()+errBuf.String(), tc.want) {
				t.Fatalf("output %q %q does not contain %q", out.String(), errBuf.String(), tc.want)
			}
		})
	}
}

func TestRunEnvThreshold(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cov.out")
	os.WriteFile(profile, []byte("mode: set\nfile.go:1.1,1.2 1 0\n"), 0o644)
	t.Setenv("COVER_THRESH", "0")
	var out, errBuf bytes.Buffer
	code := run([]string{"covercheck", profile}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errBuf.String())
	}
}
