// main.go — SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Short:        "Aligns HCL files based on given criteria",
		Args:         cobra.ArbitraryArgs,
		RunE:         cli.RunE,
		SilenceUsage: true,
	}

	rootCmd.Flags().Bool("write", false, "write result to file(s)")
	rootCmd.Flags().Bool("check", false, "check if files are formatted")
	rootCmd.Flags().Bool("diff", false, "print the diff of required changes")
	rootCmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	rootCmd.Flags().Bool("stdin", false, "read from STDIN")
	rootCmd.Flags().Bool("stdout", false, "write result to STDOUT")
	rootCmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
	rootCmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
	rootCmd.Flags().StringSlice("order", config.CanonicalOrder, "order of variable block fields")
	rootCmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
	rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	rootCmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")

	if err := rootCmd.Execute(); err != nil {
		if ec, ok := err.(*cli.ExitCodeError); ok {
			os.Exit(ec.Code)
		}
		os.Exit(1)
	}
}
