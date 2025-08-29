package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/hashicorp/hclalign/cli"
)

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
