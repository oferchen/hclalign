// cmd/commentcheck/main_test.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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
		t.Fatalf("expected error, got %v", err)
	}
}

func TestFixFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.go")
	write(t, path, "// extra\npackage main\n// trailing\n")
	if err := fixFile(path); err != nil {
		t.Fatalf("fix file: %v", err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	want := fmt.Sprintf("// %s\npackage main\n\n", filepath.ToSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != want {
		t.Fatalf("fix mismatch: %q != %q", data, want)
	}
}

func TestGoFilesGitNotFound(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(string, ...string) *exec.Cmd { return exec.Command("bash", "-c", "exit 1") }
	if _, err := goFiles(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestMainModes(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	path := filepath.Join(dir, "foo.go")
	write(t, path, "// bad\npackage main\n")

	gitOutput := fmt.Sprintf("%s\n", filepath.Base(path))
	origExec := execCommand
	origExit := osExit
	defer func() { execCommand = origExec; osExit = origExit }()
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "git" && len(args) > 0 && args[0] == "ls-files" {
			return exec.Command("bash", "-c", fmt.Sprintf("printf '%s'", gitOutput))
		}
		return exec.Command("bash", "-c", "")
	}
	var code int
	osExit = func(c int) { code = c }
	os.Args = []string{"cmd", "--mode=fix"}
	main()
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := fmt.Sprintf("// %s\npackage main\n\n", filepath.ToSlash(filepath.Base(path)))
	if string(data) != want {
		t.Fatalf("fix via main failed: %q", data)
	}

       code = 0
       os.Args = []string{"cmd", "--mode=ci"}
       main()
       if code != 0 {
		t.Fatalf("unexpected exit %d", code)
	}
}
