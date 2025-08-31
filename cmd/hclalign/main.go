// cmd/hclalign/main.go
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
		Use:   "hclalign [target file or directory]",
		Short: "Aligns HCL files based on given criteria",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return &cli.ExitCodeError{Err: err, Code: 2}
			}

			count := 0
			for _, f := range []string{"write", "check", "diff"} {
				if cmd.Flags().Changed(f) {
					count++
				}
			}
			if count > 1 {
				return &cli.ExitCodeError{Err: errors.New("write, check, and diff are mutually exclusive"), Code: 2}
			}

			count = 0
			for _, f := range []string{"types", "all"} {
				if cmd.Flags().Changed(f) {
					count++
				}
			}
			if count > 1 {
				return &cli.ExitCodeError{Err: errors.New("types and all are mutually exclusive"), Code: 2}
			}

			return nil
		},
		RunE:         runE,
		SilenceUsage: true,
	}

	rootCmd.Flags().Bool("write", true, "write result to files")
	rootCmd.Flags().Bool("check", false, "check if files are formatted")
	rootCmd.Flags().Bool("diff", false, "print the diff of required changes")
	rootCmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	rootCmd.Flags().Bool("stdin", false, "read from STDIN")
	rootCmd.Flags().Bool("stdout", false, "write result to STDOUT")
	rootCmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
	rootCmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
	rootCmd.Flags().Bool("follow-symlinks", false, "follow symbolic links when traversing directories")
	rootCmd.Flags().String("providers-schema", "", "path to providers schema file")
	rootCmd.Flags().Bool("use-terraform-schema", false, "use terraform schema for providers")
	rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	rootCmd.Flags().StringSlice("types", []string{"variable"}, "comma-separated list of block types to align")
	rootCmd.Flags().Bool("all", false, "align all block types")
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &cli.ExitCodeError{Err: err, Code: 2}
	})

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
