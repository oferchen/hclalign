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
=======

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


	"errors"
	"os/exec"
	"testing"
)


func TestPackageDirs_GoNotFound(t *testing.T) {
	originalLookPath := lookPath
	t.Cleanup(func() { lookPath = originalLookPath })
	lookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}
	if _, err := packageDirs(); err == nil || !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("expected exec.ErrNotFound, got %v", err)
	}
}

func TestPackageDirs_CommandFailure(t *testing.T) {
	originalLookPath := lookPath
	originalExecCommand := execCommand
	t.Cleanup(func() {
		lookPath = originalLookPath
		execCommand = originalExecCommand
	})
	lookPath = func(string) (string, error) { return "go", nil }
	execCommand = func(string, ...string) *exec.Cmd {
		return exec.Command("bash", "-c", "exit 1")
	}
	if _, err := packageDirs(); err == nil {
		t.Fatalf("expected error, got nil")
	}
=======
func TestPackageDirsNoGoBinary(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("nonexistent-go-binary")
	}
	_, err := packageDirs()
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "commentcheck requires a Go toolchain" {

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
=======
func TestPackageDirsCommandError(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "exit 1")
	}
	_, err := packageDirs()
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() == "commentcheck requires a Go toolchain" {
		t.Fatalf("unexpected error: %v", err)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}

}
