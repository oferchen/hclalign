// cli/cli.go
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
	cfg, err := parseConfig(cmd, args)
	if err != nil {
		return err
	}

	changed, err := engine.Process(cmd.Context(), cfg)
	if err != nil {
		return &ExitCodeError{Err: err, Code: 3}
	}

	if changed && (cfg.Mode == config.ModeCheck || cfg.Mode == config.ModeDiff) {
		return &ExitCodeError{Err: fmt.Errorf("files need formatting"), Code: 1}
	}

	return nil
}
