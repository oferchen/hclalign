// main.go

package main

import (
	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hclalign [target file or directory]",
		Short: "Aligns HCL files based on given criteria",
		Args:  cobra.MinimumNArgs(1),
		RunE:  cli.RunE,
	}

	rootCmd.Flags().StringSliceP("criteria", "c", config.DefaultCriteria, "List of file criteria to align")
	rootCmd.Flags().StringSliceP("order", "o", config.DefaultOrder, "Comma-separated list of the order of variable block fields")

	cobra.CheckErr(rootCmd.Execute())
}
