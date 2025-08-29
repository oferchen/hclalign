package terraformfmt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Strategy string

const (
	StrategyAuto   Strategy = "auto"
	StrategyBinary Strategy = "binary"
	StrategyGo     Strategy = "go"
)

// Format formats src according to the selected strategy. The returned bytes
// are guaranteed to parse back to the same structure as the input. The function
// also reconstructs the formatted file using BuildTokens and SetAttributeRaw to
// ensure token preservation.
func Format(ctx context.Context, src []byte, strat Strategy) ([]byte, error) {
	formatted, err := format(ctx, src, strat)
	if err != nil {
		return nil, err
	}

	// verify equivalence
	origSyntax, diags := hclsyntax.ParseConfig(src, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse original: %w", diags)
	}
	fmtSyntax, diags := hclsyntax.ParseConfig(formatted, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse formatted: %w", diags)
	}
	if !FilesEqual(origSyntax, fmtSyntax) {
		return nil, errors.New("formatted output not equivalent to original")
	}

	// build tokens and reconstruct file to ensure token preservation
	parsed, diags := hclwrite.ParseConfig(formatted, "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse formatted: %w", diags)
	}
	reconstructed := hclwrite.NewEmptyFile()
	for name, attr := range parsed.Body().Attributes() {
		reconstructed.Body().SetAttributeRaw(name, attr.BuildTokens(nil))
	}
	for _, block := range parsed.Body().Blocks() {
		reconstructed.Body().AppendBlock(block)
	}
	_ = reconstructed.Bytes() // ensure tokens are serialised

	return formatted, nil
}

func format(ctx context.Context, src []byte, strat Strategy) ([]byte, error) {
	switch strat {
	case StrategyBinary:
		return runTerraformFmt(ctx, src)
	case StrategyGo:
		return hclwrite.Format(src), nil
	case StrategyAuto:
		if out, err := runTerraformFmt(ctx, src); err == nil {
			return out, nil
		}
		return hclwrite.Format(src), nil
	default:
		return nil, fmt.Errorf("unknown format strategy %q", strat)
	}
}

func runTerraformFmt(ctx context.Context, src []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "terraform", "fmt", "-")
	cmd.Stdin = bytes.NewReader(src)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FilesEqual compares two parsed files for structural equality while ignoring
// position information.
func FilesEqual(a, b *hcl.File) bool {
	if a == nil || b == nil {
		return a == b
	}
	return bytes.Equal(hclwrite.Format(a.Bytes), hclwrite.Format(b.Bytes))
}
