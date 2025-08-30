// internal/align/reorder.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

func reorderBlock(block *hclwrite.Block, order []string) error {
	body := block.Body()
	attrs := body.Attributes()
	nestedBlocks := body.Blocks()

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	attrTokensMap := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokensMap[name] = ihcl.ExtractAttrTokens(attr)
	}

	for name := range attrs {
		body.RemoveAttribute(name)
	}
	for _, nb := range nestedBlocks {
		body.RemoveBlock(nb)
	}

	body.Clear()
	if len(order) > 0 || len(nestedBlocks) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}
	for _, name := range order {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}
	for _, nb := range nestedBlocks {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
		body.AppendBlock(nb)
	}
	if trailingComma && (len(order) > 0 || len(nestedBlocks) > 0) {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
		})
	}
	toks := body.BuildTokens(nil)
	if len(toks) > 0 && toks[len(toks)-1].Type != hclsyntax.TokenNewline {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}
	return nil
}
