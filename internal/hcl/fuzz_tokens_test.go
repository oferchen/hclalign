package hcl

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func FuzzExtractAttrTokens(f *testing.F) {
	f.Add("1")
	f.Fuzz(func(t *testing.T, expr string) {
		src := []byte(fmt.Sprintf("a = %s\n", expr))
		file, diags := hclwrite.ParseConfig(src, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Skip()
		}
		attr := file.Body().GetAttribute("a")
		toks := ExtractAttrTokens(attr)
		body := hclwrite.NewEmptyFile().Body()
		raw := append(hclwrite.Tokens{}, toks.LeadTokens...)
		raw = append(raw, toks.ExprTokens...)
		body.SetAttributeRaw("a", raw)
		out := hclwrite.Format(body.BuildTokens(nil).Bytes())
		_, diags = hclwrite.ParseConfig(out, "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("round-trip parse: %v", diags)
		}
	})
}

func FuzzNormalizeTokens(f *testing.F) {
	f.Add("a = 1\n")
	f.Fuzz(func(t *testing.T, src string) {
		file, diags := hclwrite.ParseConfig([]byte(src), "fuzz.hcl", hcl.InitialPos)
		if diags.HasErrors() {
			t.Skip()
		}
		tokens := file.Body().BuildTokens(nil)
		NormalizeTokens(tokens)
	})
}
