package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func writeProfile(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile("coverage.out", []byte(content), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	t.Cleanup(func() { os.Remove("coverage.out") })
}

func TestCoverageAboveThreshold(t *testing.T) {
	writeProfile(t, "mode: set\nfoo.go:1.1,1.10 1 1\n")
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Total coverage: 100.0%") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCoverageBelowThreshold(t *testing.T) {
	writeProfile(t, "mode: set\nfoo.go:1.1,1.10 1 0\n")
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error, got none: %s", out)
	}
	if !strings.Contains(string(out), "Coverage 0.0% is below 95.0%") {
		t.Fatalf("unexpected output: %s", out)
	}
}
