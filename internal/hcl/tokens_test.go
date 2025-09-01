// internal/hcl/tokens_test.go
package hcl

import (
	"bytes"
	"testing"

	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
)

func TestExtractAttrTokens(t *testing.T) {
	src := []byte("//c\nname = 1 //trail\n")
	f, _ := hclwrite.ParseConfig(src, "test.tf", hcl2.InitialPos)
	body := f.Body()
	attr := body.GetAttribute("name")
	tokensMap, start := ExtractAttrTokens(body, map[string]*hclwrite.Attribute{"name": attr})
	require.Len(t, start, 0)
	toks := tokensMap["name"]
	require.Len(t, toks.PreTokens, 0)
	require.Len(t, toks.LeadTokens, 1)
	require.Equal(t, "//c\n", string(toks.LeadTokens[0].Bytes))
	var buf bytes.Buffer
	for _, tk := range toks.ExprTokens {
		buf.Write(tk.Bytes)
	}
	require.Equal(t, "1//trail", buf.String())
	require.Len(t, toks.PostTokens, 0)
}

func TestExtractAttrTokensPrePost(t *testing.T) {
	src := []byte("a=1\n#ta\n\nb=2//ib\n#tb\n")
	f, _ := hclwrite.ParseConfig(src, "t.tf", hcl2.InitialPos)
	body := f.Body()
	attrs := map[string]*hclwrite.Attribute{
		"a": body.GetAttribute("a"),
		"b": body.GetAttribute("b"),
	}
	toks, start := ExtractAttrTokens(body, attrs)
	require.Len(t, start, 0)
	a := toks["a"]
	require.Len(t, a.PostTokens, 1)
	require.Equal(t, "#ta\n", string(a.PostTokens[0].Bytes))
	b := toks["b"]
	require.Len(t, b.PreTokens, 1)
	require.Equal(t, "\n", string(b.PreTokens[0].Bytes))
	require.Len(t, b.PostTokens, 1)
	require.Equal(t, "#tb\n", string(b.PostTokens[0].Bytes))
}

func TestHasTrailingComma(t *testing.T) {
	src := []byte("a = {b = 1,}\nb = {c = 2}\n")
	f, _ := hclwrite.ParseConfig(src, "t.tf", hcl2.InitialPos)
	attr1 := f.Body().GetAttribute("a")
	require.True(t, HasTrailingComma(attr1.Expr().BuildTokens(nil)))
	attr2 := f.Body().GetAttribute("b")
	require.False(t, HasTrailingComma(attr2.Expr().BuildTokens(nil)))
}

func TestDetectLineEnding(t *testing.T) {
	tokens := hclwrite.Tokens{&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\r\n")}}
	require.Equal(t, []byte("\r\n"), DetectLineEnding(tokens))
	tokens[0].Bytes = []byte("\n")
	require.Equal(t, []byte("\n"), DetectLineEnding(tokens))
}

func TestNormalizeTokens(t *testing.T) {
	tokens := hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenComment, Bytes: append([]byte{0xEF, 0xBB, 0xBF}, []byte("line\r\n")...)},
	}
	bom := NormalizeTokens(tokens)
	require.True(t, bom)
	require.Equal(t, "line\n", string(tokens[0].Bytes))
}

func TestAttributeOrder(t *testing.T) {
	src := []byte("b = 1\na = {x = 1}\nblock {\n d = 1\n}\nd = 2\n")
	f, _ := hclwrite.ParseConfig(src, "t.tf", hcl2.InitialPos)
	attrs := map[string]*hclwrite.Attribute{
		"a": f.Body().GetAttribute("a"),
		"b": f.Body().GetAttribute("b"),
		"d": f.Body().GetAttribute("d"),
	}
	order := AttributeOrder(f.Body(), attrs)
	require.Equal(t, []string{"b", "a", "d"}, order)
}
