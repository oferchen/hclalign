// main.go

package main

import (
	"log"
	"os"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hcl_align [target file or directory]",
		Short: "Aligns HCL files based on given criteria",
		//Args:  cobra.ExactArgs(1),
		Args: cobra.NoArgs,
		RunE: cli.RunE,
	}

	rootCmd.Flags().StringSliceP("criteria", "c", config.DefaultCriteria, "List of file criteria to align")
	rootCmd.Flags().StringSliceP("order", "o", config.DefaultOrder, "Comma-separated list of the order of variable block fields")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}
