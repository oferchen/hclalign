// /cmd/hclalign/main_test.go
package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/cli"
)

func TestMainExitCode(t *testing.T) {
	oldOsExit := osExit
	oldArgs := os.Args
	oldRunE := runE
	t.Cleanup(func() {
		osExit = oldOsExit
		os.Args = oldArgs
		runE = oldRunE
	})

	var gotCode int
	osExit = func(code int) { gotCode = code }

	runE = func(_ *cobra.Command, _ []string) error {
		return &cli.ExitCodeError{Err: errors.New("boom"), Code: 42}
	}

	os.Args = []string{"hclalign"}
	main()

	if gotCode != 42 {
		t.Fatalf("expected exit code 42, got %d", gotCode)
	}
}

func TestRunCheckFlag(t *testing.T) {
	oldRunE := runE
	t.Cleanup(func() { runE = oldRunE })

	runE = func(cmd *cobra.Command, args []string) error {
		check, err := cmd.Flags().GetBool("check")
		if err != nil {
			t.Fatalf("get flag: %v", err)
		}
		if !check {
			t.Fatalf("expected check flag to be true")
		}
		diff, err := cmd.Flags().GetBool("diff")
		if err != nil {
			t.Fatalf("get flag: %v", err)
		}
		if diff {
			t.Fatalf("expected diff flag to be false")
		}
		return nil
	}

	if code := run([]string{"--check"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunMutuallyExclusiveFlags(t *testing.T) {
	oldRunE := runE
	t.Cleanup(func() { runE = oldRunE })

	runE = func(_ *cobra.Command, _ []string) error {
		t.Fatal("runE should not be called")
		return nil
	}

	if code := run([]string{"--check", "--diff"}); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func TestRunWrappedExitCode(t *testing.T) {
	oldRunE := runE
	t.Cleanup(func() { runE = oldRunE })

	runE = func(_ *cobra.Command, _ []string) error {
		return fmt.Errorf("wrap: %w", &cli.ExitCodeError{Err: errors.New("boom"), Code: 7})
	}

	if code := run(nil); code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
}
