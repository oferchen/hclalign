// internal/hcl/fuzz_tokens_test.go
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
		toksMap, start := ExtractAttrTokens(file.Body(), map[string]*hclwrite.Attribute{"a": attr})
		toks := toksMap["a"]
		body := hclwrite.NewEmptyFile().Body()
		body.AppendUnstructuredTokens(start)
		raw := append(hclwrite.Tokens{}, toks.LeadTokens...)
		raw = append(raw, toks.ExprTokens...)
		body.AppendUnstructuredTokens(toks.PreTokens)
		body.SetAttributeRaw("a", raw)
		body.AppendUnstructuredTokens(toks.PostTokens)
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
