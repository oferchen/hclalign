// cmd/commentcheck/main_test.go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		t.Fatalf("check file: %v", err)
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
}

func TestMainSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	write(t, path, "")

	originalPackageDirs := packageDirs
	originalCheckFile := checkFile
	originalExit := osExit
	t.Cleanup(func() {
		packageDirs = originalPackageDirs
		checkFile = originalCheckFile
		osExit = originalExit
	})

	packageDirs = func() ([]string, error) { return []string{dir}, nil }
	checkFile = func(f string) error {
		if f != path {
			t.Fatalf("unexpected path %s", f)
		}
		return nil
	}

	var code = -1
	osExit = func(c int) { code = c }
	main()
	if code != -1 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestMainFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	write(t, path, "")

	originalPackageDirs := packageDirs
	originalCheckFile := checkFile
	originalExit := osExit
	t.Cleanup(func() {
		packageDirs = originalPackageDirs
		checkFile = originalCheckFile
		osExit = originalExit
	})

	packageDirs = func() ([]string, error) { return []string{dir}, nil }
	checkFile = func(string) error { return errors.New("fail") }

	var code = -1
	osExit = func(c int) { code = c }
	main()
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
