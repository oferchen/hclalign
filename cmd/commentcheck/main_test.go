package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckFile(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.go")
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("wd: %v", err)
		}
		rel, err := filepath.Rel(wd, path)
		if err != nil {
			t.Fatalf("rel: %v", err)
		}
		content := "// " + filepath.ToSlash(rel) + "\npackage main\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := checkFile(rel); err != nil {
			t.Fatalf("checkFile returned error: %v", err)
		}
	})

	t.Run("missing comment", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.go")
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("wd: %v", err)
		}
		rel, err := filepath.Rel(wd, path)
		if err != nil {
			t.Fatalf("rel: %v", err)
		}
		// write file without leading comment
		if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := checkFile(rel); err == nil {
			t.Fatal("expected error for missing comment")
		}
	})

	t.Run("wrong comment", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.go")
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("wd: %v", err)
		}
		rel, err := filepath.Rel(wd, path)
		if err != nil {
			t.Fatalf("rel: %v", err)
		}
		content := "// wrong\npackage main\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := checkFile(rel); err == nil {
			t.Fatal("expected error for wrong comment")
		}
	})
}
