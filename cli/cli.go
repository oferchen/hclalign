// cli.go
// Defines the CLI interface and command execution logic.

package cli

import (
	"fmt"

	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

func RunE(cmd *cobra.Command, args []string) error {
	target := args[0]
	criteria, err := cmd.Flags().GetStringSlice("criteria")
	if err != nil {
		return err
	}
	order, err := cmd.Flags().GetStringSlice("order")
	if err != nil {
		return err
	}

	if _, err := config.IsValidOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	return config.ProcessTargetDynamically(cmd.Context(), target, criteria, order)
}
