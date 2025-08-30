// internal/ci/covercheck/main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const threshold = 95.0

var ignore = []string{
	"internal/ci/",
	"main.go",
}

func run(args []string, out, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(errOut, "usage: covercheck <profile>")
		return 1
	}
	profile := args[1]

	f, err := os.Open(profile)
	if err != nil {
		fmt.Fprintf(errOut, "could not open %s: %v\n", profile, err)
		return 1
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if !s.Scan() {
		fmt.Fprintln(errOut, "empty coverage profile")
		return 1
	}

	var total, covered float64
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 3 {
			continue
		}
		file := fields[0]
		skip := false
		for _, p := range ignore {
			if strings.Contains(file, p) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		stmts, err1 := strconv.ParseFloat(fields[1], 64)
		count, err2 := strconv.ParseFloat(fields[2], 64)
		if err1 != nil || err2 != nil {
			fmt.Fprintf(errOut, "invalid line: %s\n", s.Text())
			return 1
		}
		total += stmts
		if count > 0 {
			covered += stmts
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(errOut, "error reading coverage profile: %v\n", err)
		return 1
	}

	if total == 0 {
		fmt.Fprintln(errOut, "no statements in coverage profile")
		return 1
	}

	pct := covered / total * 100
	fmt.Fprintf(out, "Total coverage: %.1f%%\n", pct)
	if pct < threshold {
		fmt.Fprintf(errOut, "Coverage %.1f%% is below %.1f%%\n", pct, threshold)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
