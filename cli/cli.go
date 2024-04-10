// cli.go
// Defines the CLI interface and command execution logic.

package cli

import (
	"fmt"

	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

func RunE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Printf("Error: accepts 1 arg(s), received %d\n\n", len(args))
		cmd.Usage()
		return fmt.Errorf(config.MissingTarget)
	}
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
	orderValid, err := config.IsValidOrder(order)
	if !orderValid {
		return fmt.Errorf("invalid order provided: %v", err)
	}
	// Process target dynamically
	return config.ProcessTargetDynamically(target, criteria, order)
}
