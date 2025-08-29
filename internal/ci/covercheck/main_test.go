// internal/ci/covercheck/main_test.go — SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runCovercheck(t *testing.T) (string, int) {
	t.Helper()
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("unexpected error: %v", err)
	return "", 0
}

func TestCovercheck(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		wantExit int
		wantMsg  string
	}{
		{
			name:     "above",
			profile:  "mode: set\nfile.go:1.1,1.2 1 1\n",
			wantExit: 0,
			wantMsg:  "Total coverage: 100.0%",
		},
		{
			name:     "below",
			profile:  "mode: set\nfile.go:1.1,1.2 1 0\nfile.go:2.1,2.2 1 1\n",
			wantExit: 1,
			wantMsg:  "Coverage 50.0% is below 95.0%",
		},
		{
			name:     "ignored path",
			profile:  "mode: set\ncmd/commentcheck/file.go:1.1,1.2 1 0\nfile.go:2.1,2.2 1 1\n",
			wantExit: 0,
			wantMsg:  "Total coverage: 100.0%",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile("coverage.out", []byte(tc.profile), 0644); err != nil {
				t.Fatalf("write profile: %v", err)
			}
			defer os.Remove("coverage.out")
			out, code := runCovercheck(t)
			if code != tc.wantExit {
				t.Fatalf("exit %d, want %d; output: %s", code, tc.wantExit, out)
			}
			if !strings.Contains(out, tc.wantMsg) {
				t.Fatalf("output %q does not contain %q", out, tc.wantMsg)
			}
		})
	}
}
