// cli/parse.go
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

	var err error
	writeMode := getBool(cmd, "write", &err)
	checkMode := getBool(cmd, "check", &err)
	diffMode := getBool(cmd, "diff", &err)
	stdin := getBool(cmd, "stdin", &err)
	stdout := getBool(cmd, "stdout", &err)
	include := getStringSlice(cmd, "include", &err)
	exclude := getStringSlice(cmd, "exclude", &err)
	orderRaw := getStringSlice(cmd, "order", &err)
	prefixOrder := getBool(cmd, "prefix-order", &err)
	providersSchema := getString(cmd, "providers-schema", &err)
	useTerraformSchema := getBool(cmd, "use-terraform-schema", &err)
	concurrency := getInt(cmd, "concurrency", &err)
	verbose := getBool(cmd, "verbose", &err)
	types := getStringSlice(cmd, "types", &err)
	all := getBool(cmd, "all", &err)
	if err != nil {
		return nil, err
	}

	if (writeMode && checkMode) || (writeMode && diffMode) || (checkMode && diffMode) {
		return nil, &ExitCodeError{Err: fmt.Errorf("cannot specify more than one of --write, --check, or --diff"), Code: 2}
	}

	attrOrder, err := config.ParseOrder(orderRaw)
	if err != nil {
		return nil, &ExitCodeError{Err: err, Code: 2}
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
		PrefixOrder:        prefixOrder,
		ProvidersSchema:    providersSchema,
		UseTerraformSchema: useTerraformSchema,
		Concurrency:        concurrency,
		Verbose:            verbose,
		Types:              cfgTypes,
	}

	if err := cfg.Validate(); err != nil {
		return nil, &ExitCodeError{Err: err, Code: 2}
	}

	return cfg, nil
}

func getBool(cmd *cobra.Command, name string, err *error) bool {
	if *err != nil {
		return false
	}
	var v bool
	v, *err = cmd.Flags().GetBool(name)
	if *err != nil {
		*err = fmt.Errorf("get flag %s: %w", name, *err)
	}
	return v
}

func getStringSlice(cmd *cobra.Command, name string, err *error) []string {
	if *err != nil {
		return nil
	}
	var v []string
	v, *err = cmd.Flags().GetStringSlice(name)
	if *err != nil {
		*err = fmt.Errorf("get flag %s: %w", name, *err)
	}
	return v
}

func getString(cmd *cobra.Command, name string, err *error) string {
	if *err != nil {
		return ""
	}
	var v string
	v, *err = cmd.Flags().GetString(name)
	if *err != nil {
		*err = fmt.Errorf("get flag %s: %w", name, *err)
	}
	return v
}

func getInt(cmd *cobra.Command, name string, err *error) int {
	if *err != nil {
		return 0
	}
	var v int
	v, *err = cmd.Flags().GetInt(name)
	if *err != nil {
		*err = fmt.Errorf("get flag %s: %w", name, *err)
	}
	return v
}
