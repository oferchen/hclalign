// /cli/parse.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/config"
)

func parseConfig(cmd *cobra.Command, args []string) (*config.Config, error) {
	var target string
	if len(args) == 1 {
		target = args[0]
	}

	writeMode, err := cmd.Flags().GetBool("write")
	if err != nil {
		return nil, err
	}
	checkMode, err := cmd.Flags().GetBool("check")
	if err != nil {
		return nil, err
	}
	diffMode, err := cmd.Flags().GetBool("diff")
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.Flags().GetBool("stdin")
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.Flags().GetBool("stdout")
	if err != nil {
		return nil, err
	}
	include, err := cmd.Flags().GetStringSlice("include")
	if err != nil {
		return nil, err
	}
	exclude, err := cmd.Flags().GetStringSlice("exclude")
	if err != nil {
		return nil, err
	}
	orderRaw, err := cmd.Flags().GetStringSlice("order")
	if err != nil {
		return nil, err
	}
	attrOrder, blockOrder, err := config.ParseOrder(orderRaw)
	if err != nil {
		return nil, &ExitCodeError{Err: err, Code: 2}
	}
	fmtOnly, err := cmd.Flags().GetBool("fmt-only")
	if err != nil {
		return nil, err
	}
	noFmt, err := cmd.Flags().GetBool("no-fmt")
	if err != nil {
		return nil, err
	}
	fmtStrategy, err := cmd.Flags().GetString("fmt-strategy")
	if err != nil {
		return nil, err
	}
	providersSchema, err := cmd.Flags().GetString("providers-schema")
	if err != nil {
		return nil, err
	}
	useTerraformSchema, err := cmd.Flags().GetBool("use-terraform-schema")
	if err != nil {
		return nil, err
	}
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return nil, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	followSymlinks, err := cmd.Flags().GetBool("follow-symlinks")
	if err != nil {
		return nil, err
	}
	types, err := cmd.Flags().GetStringSlice("types")
	if err != nil {
		return nil, err
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return nil, err
	}
	sortUnknown, err := cmd.Flags().GetBool("sort-unknown")
	if err != nil {
		return nil, err
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
		return nil, &ExitCodeError{Err: fmt.Errorf("cannot specify more than one of --write, --check, or --diff"), Code: 2}
	}
	if fmtOnly && noFmt {
		return nil, &ExitCodeError{Err: fmt.Errorf("cannot specify both --fmt-only and --no-fmt"), Code: 2}
	}

	if !stdin && target == "" {
		return nil, &ExitCodeError{Err: fmt.Errorf(config.ErrMissingTarget), Code: 2}
	}
	if stdin && target != "" {
		return nil, &ExitCodeError{Err: fmt.Errorf("cannot specify target when --stdin is used"), Code: 2}
	}
	if stdin && !stdout {
		return nil, &ExitCodeError{Err: fmt.Errorf("--stdout is required when --stdin is used"), Code: 2}
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

	var cfgTypes []string
	if all {
		cfgTypes = nil
	} else {
		cfgTypes = types
	}

	cfg := &config.Config{
		Target:             target,
		Mode:               mode,
		Stdin:              stdin,
		Stdout:             stdout,
		Include:            include,
		Exclude:            exclude,
		Order:              attrOrder,
		BlockOrder:         blockOrder,
		FmtOnly:            fmtOnly,
		NoFmt:              noFmt,
		FmtStrategy:        fmtStrategy,
		ProvidersSchema:    providersSchema,
		UseTerraformSchema: useTerraformSchema,
		Concurrency:        concurrency,
		Verbose:            verbose,
		FollowSymlinks:     followSymlinks,
		Types:              cfgTypes,
		SortUnknown:        sortUnknown,
	}

	if err := cfg.Validate(); err != nil {
		return nil, &ExitCodeError{Err: err, Code: 2}
	}

	return cfg, nil
}
