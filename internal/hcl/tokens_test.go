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
	attr := f.Body().GetAttribute("name")
	toks := ExtractAttrTokens(attr)
	require.Len(t, toks.LeadTokens, 1)
	require.Equal(t, "//c\n", string(toks.LeadTokens[0].Bytes))
	var buf bytes.Buffer
	for _, tk := range toks.ExprTokens {
		buf.Write(tk.Bytes)
	}
	require.Equal(t, "1//trail", buf.String())
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
