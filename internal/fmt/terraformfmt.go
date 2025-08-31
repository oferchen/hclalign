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

func Format(ctx context.Context, src []byte, filename, strategy string) ([]byte, internalfs.Hints, error) {
	if err := ctx.Err(); err != nil {
		return nil, internalfs.Hints{}, err
	}
	switch Strategy(strategy) {
	case StrategyGo:
		return formatter.Format(src, filename)
	case StrategyBinary:
		return formatBinary(ctx, src)
	case StrategyAuto, "":
		return Run(ctx, src)
	default:
		return nil, internalfs.Hints{}, fmt.Errorf("unknown fmt strategy %q", strategy)
	}
}

func FormatFile(ctx context.Context, path string) ([]byte, internalfs.Hints, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, internalfs.Hints{}, false, err
	}
	if _, err := exec.LookPath("terraform"); err != nil {
		return nil, internalfs.Hints{}, false, nil
	}
	data, _, hints, err := internalfs.ReadFileWithHints(ctx, path)
	if err != nil {
		return nil, hints, false, err
	}
	prepared := internalfs.PrepareForParse(data, hints)
	if len(prepared) > 0 && !utf8.Valid(prepared) {
		return nil, hints, false, fmt.Errorf("input is not valid UTF-8")
	}
	cmd := exec.CommandContext(ctx, "terraform", "fmt", "-no-color", "-list=false", "-write=false", path)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderrStr := stderr.String(); stderrStr != "" {
			return nil, hints, false, fmt.Errorf("terraform fmt failed: %w: %s", err, stderrStr)
		}
		return nil, hints, false, fmt.Errorf("terraform fmt failed: %w", err)
	}
	formatted := stdout.Bytes()
	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}
	return formatted, hints, true, nil
}

func formatBinary(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
	hints := internalfs.DetectHintsFromBytes(src)
	src = internalfs.PrepareForParse(src, hints)
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, hints, fmt.Errorf("input is not valid UTF-8")
	}
	path := terraformBinary()
	if path == "" {
		return nil, hints, fmt.Errorf("terraform binary not found")
	}
	cmd := exec.CommandContext(ctx, path, "fmt", "-no-color", "-list=false", "-write=false", "-")
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
