// main.go

package main

import (
	"github.com/oferchen/hclalign/cli"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hcl_align [target file or directory]",
		Short: "Aligns HCL files based on given criteria",
		//Args:  cobra.ExactArgs(1),
		Args: cobra.NoArgs,
		RunE: cli.RunE,
	}

	rootCmd.Flags().StringSliceP("criteria", "c", cli.DefaultCriteria, "List of file criteria to align")
	rootCmd.Flags().StringSliceP("order", "o", cli.DefaultOrder, "Comma-separated list of the order of variable block fields")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}
