// cli.go
// Defines the CLI interface and command execution logic.

package cli

import (
	"fmt"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

// DefaultCriteria and DefaultOrder define the default behavior of the CLI.
var (
	DefaultCriteria = []string{"*.tf"}
	DefaultOrder    = []string{"description", "type", "default", "sensitive", "nullable", "validation"}
)

// RunE is the execution logic for the root command.
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

	// Validate the order
	if !config.IsValidOrder(order) {
		return fmt.Errorf("invalid order: %v", order)
	}

	// Process target dynamically
	return config.ProcessTargetDynamically(target, criteria, order)
}
