// cli/cli.go — SPDX-License-Identifier: Apache-2.0
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/internal/engine"
)

type ExitCodeError struct {
	Err  error
	Code int
}

func (e *ExitCodeError) Error() string { return e.Err.Error() }

func RunE(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return &ExitCodeError{Err: fmt.Errorf("accepts at most 1 arg(s), received %d", len(args)), Code: 2}
	}

	target := ""
	if len(args) == 1 {
		target = args[0]
	}

	writeMode, err := cmd.Flags().GetBool("write")
	if err != nil {
		return err
	}
	checkMode, err := cmd.Flags().GetBool("check")
	if err != nil {
		return err
	}
	diffMode, err := cmd.Flags().GetBool("diff")
	if err != nil {
		return err
	}
	stdin, err := cmd.Flags().GetBool("stdin")
	if err != nil {
		return err
	}
	stdout, err := cmd.Flags().GetBool("stdout")
	if err != nil {
		return err
	}
	include, err := cmd.Flags().GetStringSlice("include")
	if err != nil {
		return err
	}
	exclude, err := cmd.Flags().GetStringSlice("exclude")
	if err != nil {
		return err
	}
	order, err := cmd.Flags().GetStringSlice("order")
	if err != nil {
		return err
	}
	strictOrder, err := cmd.Flags().GetBool("strict-order")
	if err != nil {
		return err
	}
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	followSymlinks, err := cmd.Flags().GetBool("follow-symlinks")
	if err != nil {
		return err
	}

	modeCount := 0
	if writeMode {
		modeCount++
	}
	if checkMode {
		modeCount++
	}
	if diffMode {
		modeCount++
	}
	if modeCount > 1 {
		return &ExitCodeError{Err: fmt.Errorf("cannot specify more than one of --write, --check, or --diff"), Code: 2}
	}

	if !stdin && target == "" {
		return &ExitCodeError{Err: fmt.Errorf(config.ErrMissingTarget), Code: 2}
	}
	if stdin && target != "" {
		return &ExitCodeError{Err: fmt.Errorf("cannot specify target when --stdin is used"), Code: 2}
	}
	if stdin && !stdout {
		return &ExitCodeError{Err: fmt.Errorf("--stdout is required when --stdin is used"), Code: 2}
	}

	var mode config.Mode
	switch {
	case writeMode:
		mode = config.ModeWrite
	case checkMode:
		mode = config.ModeCheck
	case diffMode:
		mode = config.ModeDiff
	default:
		mode = config.ModeWrite
	}

	cfg := &config.Config{
		Target:         target,
		Mode:           mode,
		Stdin:          stdin,
		Stdout:         stdout,
		Include:        include,
		Exclude:        exclude,
		Order:          order,
		StrictOrder:    strictOrder,
		Concurrency:    concurrency,
		Verbose:        verbose,
		FollowSymlinks: followSymlinks,
	}

	if err := cfg.Validate(); err != nil {
		return &ExitCodeError{Err: err, Code: 2}
	}

	changed, err := engine.Process(cmd.Context(), cfg)
	if err != nil {
		return &ExitCodeError{Err: err, Code: 3}
	}

	if changed && (mode == config.ModeCheck || mode == config.ModeDiff) {
		return &ExitCodeError{Err: fmt.Errorf("files need formatting"), Code: 1}
	}

	return nil
}
