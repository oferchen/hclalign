// internal/fmt/terraformfmt.go
package terraformfmt

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"unicode/utf8"

	"github.com/oferchen/hclalign/formatter"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

type Strategy string

const (
	StrategyAuto   Strategy = "auto"
	StrategyBinary Strategy = "binary"
	StrategyGo     Strategy = "go"
)

func Format(src []byte, filename, strategy string) ([]byte, internalfs.Hints, error) {
	switch Strategy(strategy) {
	case StrategyGo:
		return formatter.Format(src, filename)
	case StrategyBinary:
		return formatBinary(context.Background(), src)
	case StrategyAuto, "":
		return Run(context.Background(), src)
	default:
		return nil, internalfs.Hints{}, fmt.Errorf("unknown fmt strategy %q", strategy)
	}
}

func formatBinary(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
	hints := internalfs.DetectHintsFromBytes(src)
	src = internalfs.PrepareForParse(src, hints)
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, hints, fmt.Errorf("input is not valid UTF-8")
	}
	cmd := exec.CommandContext(ctx, "terraform", "fmt", "-no-color", "-list=false", "-write=false", "-")
	cmd.Stdin = bytes.NewReader(src)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderrStr := stderr.String(); stderrStr != "" {
			return nil, hints, fmt.Errorf("terraform fmt failed: %w: %s", err, stderrStr)
		}
		return nil, hints, fmt.Errorf("terraform fmt failed: %w", err)
	}
	formatted := stdout.Bytes()

	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}
	return formatted, hints, nil
}
