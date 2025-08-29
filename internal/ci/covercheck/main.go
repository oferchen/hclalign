// internal/ci/covercheck/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const threshold = 95.0

var ignore = []string{
	"cmd/commentcheck/",
	"internal/ci/",
	"main.go",
}

func main() {
	const profile = ".build/coverage.out"

	f, err := os.Open(profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open %s: %v\n", profile, err)
		os.Exit(1)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if !s.Scan() {
		fmt.Fprintln(os.Stderr, "empty coverage profile")
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "invalid line: %s\n", s.Text())
			os.Exit(1)
		}
		total += stmts
		if count > 0 {
			covered += stmts
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading coverage profile: %v\n", err)
		os.Exit(1)
	}

	if total == 0 {
		fmt.Fprintln(os.Stderr, "no statements in coverage profile")
		os.Exit(1)
	}

	pct := covered / total * 100
	fmt.Printf("Total coverage: %.1f%%\n", pct)
	if pct < threshold {
		fmt.Fprintf(os.Stderr, "Coverage %.1f%% is below %.1f%%\n", pct, threshold)
		os.Exit(1)
	}
}

