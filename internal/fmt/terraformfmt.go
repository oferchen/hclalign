// /internal/fmt/terraformfmt.go
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

func Format(src []byte, filename, strategy string) ([]byte, error) {
	switch Strategy(strategy) {
	case StrategyGo:
		return formatter.Format(src, filename)
	case StrategyBinary:
		return formatBinary(src)
	case StrategyAuto, "":
		if _, err := exec.LookPath("terraform"); err == nil {
			return formatBinary(src)
		}
		return formatter.Format(src, filename)
	default:
		return nil, fmt.Errorf("unknown fmt strategy %q", strategy)
	}
}

func formatBinary(src []byte) ([]byte, error) {
	hints := internalfs.DetectHintsFromBytes(src)
	src = internalfs.PrepareForParse(src, hints)
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "terraform", "fmt", "-no-color", "-list=false", "-write=false", "-")
	cmd.Stdin = bytes.NewReader(src)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("terraform fmt failed: %v: %s", err, stderr.String())
	}
	formatted := stdout.Bytes()

	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}
	formatted = internalfs.ApplyHints(formatted, hints)
	return formatted, nil
}
