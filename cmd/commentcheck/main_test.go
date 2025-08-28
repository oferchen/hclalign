package main

import (
	"errors"
	"os/exec"
	"testing"
)

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
