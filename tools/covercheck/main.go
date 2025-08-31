// tools/covercheck/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	min := 95.0
	if len(os.Args) > 1 {
		if v, err := strconv.ParseFloat(os.Args[1], 64); err == nil {
			min = v
		}
	}
	scanner := bufio.NewScanner(os.Stdin)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		fmt.Fprintln(os.Stderr, "no coverage data")
		os.Exit(1)
	}
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	pct, err := strconv.ParseFloat(pctStr, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if pct < min {
		fmt.Printf("coverage %.1f%% is below %.0f%%\n", pct, min)
		os.Exit(1)
	}
}
