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

func TestCheckFileWithBuildTag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	comment := fmt.Sprintf("// %s\n", filepath.ToSlash(path))
	content := "//go:build windows\n\n" + comment + "package main\n"
	write(t, path, content)
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

func TestPackageDirs_GitNotFound(t *testing.T) {
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
	lookPath = func(string) (string, error) { return "git", nil }
	execCommand = func(string, ...string) *exec.Cmd {
		return exec.Command("bash", "-c", "exit 1")
	}
	if _, err := packageDirs(); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestPackageDirs_TestOnlyDir(t *testing.T) {
	dir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalWd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := os.Mkdir("foo", 0755); err != nil {
		t.Fatalf("mkdir foo: %v", err)
	}
	write(t, filepath.Join("foo", "bar_test.go"), "package foo\n")
	if err := exec.Command("git", "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	dirs, err := packageDirs()
	if err != nil {
		t.Fatalf("packageDirs: %v", err)
	}
	if len(dirs) != 1 || dirs[0] != "foo" {
		t.Fatalf("expected [foo], got %v", dirs)
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

