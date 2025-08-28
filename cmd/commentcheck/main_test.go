package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestCheckFileOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	comment := fmt.Sprintf("// %s\n", filepath.ToSlash(path))
	write(t, path, comment+"package main\n")
	if err := checkFile(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckFileMissingComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	write(t, path, "package main\n")
	if err := checkFile(path); err == nil || !strings.Contains(err.Error(), "first line must be") {
		t.Fatalf("expected missing comment error, got %v", err)
	}
}

func TestCheckFileMalformedComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	write(t, path, "// wrong\npackage main\n")
	if err := checkFile(path); err == nil || !strings.Contains(err.Error(), "first line must be") {
		t.Fatalf("expected malformed comment error, got %v", err)
	}
}
