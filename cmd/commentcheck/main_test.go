// /cmd/commentcheck/main_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheck(t *testing.T) {
	t.Run("compliant", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "ok.go", "//go:build test\n// ok.go\n\npackage main\n")
		if err := check(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("noncompliant extra comment", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "bad.go", "//go:build test\n// bad.go\n\npackage main\n// bad\n")
		if err := check(dir); err == nil {
			t.Fatalf("expected error")
		}
	})
	t.Run("noncompliant missing comment", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "bad.go", "//go:build test\n\npackage main\n")
		if err := check(dir); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
