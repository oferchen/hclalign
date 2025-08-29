package terraformfmt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"unicode/utf8"

	"github.com/hashicorp/hclalign/formatter"
	internalfs "github.com/hashicorp/hclalign/internal/fs"
)

type Strategy string

const (
	StrategyAuto   Strategy = "auto"
	StrategyBinary Strategy = "binary"
	StrategyGo     Strategy = "go"
)

// Format formats HCL source using the specified strategy.
// The filename parameter is used only for parse diagnostics.
func Format(src []byte, filename, strat string) ([]byte, error) {
	switch Strategy(strat) {
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
		return nil, fmt.Errorf("unknown fmt strategy %q", strat)
	}
}

func formatBinary(src []byte) ([]byte, error) {
	hints := internalfs.DetectHintsFromBytes(src)
	src = internalfs.PrepareForParse(src, hints)
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	f, err := os.CreateTemp("", "hclalign-*.tf")
	if err != nil {
		return nil, err
	}
	name := f.Name()
	if _, err := f.Write(src); err != nil {
		f.Close()
		os.Remove(name)
		return nil, err
	}
	f.Close()

	cmd := exec.Command("terraform", "fmt", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(name)
		return nil, fmt.Errorf("terraform fmt failed: %v: %s", err, string(out))
	}
	formatted, err := os.ReadFile(name)
	os.Remove(name)
	if err != nil {
		return nil, err
	}

	if len(formatted) > 0 {
		formatted = bytes.TrimRight(formatted, "\n")
		formatted = append(formatted, '\n')
	}
	formatted = internalfs.ApplyHints(formatted, hints)
	return formatted, nil
}
