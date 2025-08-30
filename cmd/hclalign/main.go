// /cmd/hclalign/main.go
package main

import (
	"errors"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
)

var (
	osExit = os.Exit
	runE   = cli.RunE
)

func main() { osExit(run(os.Args[1:])) }

func run(args []string) int {
	rootCmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Short:        "Aligns HCL files based on given criteria",
		Args:         cobra.MaximumNArgs(1),
		RunE:         runE,
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
	rootCmd.Flags().StringSlice("order", config.CanonicalOrder, "order of variable block fields and per-block ordering flags (e.g. locals=alphabetical)")
	rootCmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
	rootCmd.Flags().Bool("fmt-only", false, "only format files, skip alignment")
	rootCmd.Flags().Bool("no-fmt", false, "skip initial formatting")
	rootCmd.Flags().String("fmt-strategy", "auto", "formatting strategy to use")
	rootCmd.Flags().String("providers-schema", "", "path to providers schema file")
	rootCmd.Flags().Bool("use-terraform-schema", false, "use terraform schema for providers")
	rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	rootCmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	rootCmd.Flags().StringSlice("types", []string{"variable"}, "comma-separated list of block types to align")
	rootCmd.Flags().Bool("all", false, "align all block types")
	rootCmd.Flags().Bool("sort-unknown", false, "lexicographically sort unknown attributes")
	rootCmd.MarkFlagsMutuallyExclusive("fmt-only", "no-fmt")
	rootCmd.MarkFlagsMutuallyExclusive("types", "all")

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		var ec *cli.ExitCodeError
		if errors.As(err, &ec) {
			return ec.Code
		}
		return 1
	}
	return 0
}
