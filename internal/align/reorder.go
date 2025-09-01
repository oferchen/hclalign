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

	attrTokensMap, startTokens := ihcl.ExtractAttrTokens(body, attrs)

	for name := range attrs {
		body.RemoveAttribute(name)
	}
	for _, nb := range nestedBlocks {
		body.RemoveBlock(nb)
	}

	body.Clear()
	body.AppendUnstructuredTokens(startTokens)
	for _, name := range order {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.PreTokens)
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
			body.AppendUnstructuredTokens(tok.PostTokens)
		}
	}
	for _, nb := range nestedBlocks {
		toks := body.BuildTokens(nil)
		if len(toks) == 0 || toks[len(toks)-1].Type != hclsyntax.TokenNewline {
			body.AppendUnstructuredTokens(hclwrite.Tokens{
				&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
			})
		}
		body.AppendBlock(nb)
	}
	if trailingComma && (len(order) > 0 || len(nestedBlocks) > 0) {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
		})
	}
	toks := body.BuildTokens(nil)
	if len(toks) > 0 {
		last := toks[len(toks)-1]
		if last.Type != hclsyntax.TokenNewline {
			if last.Type != hclsyntax.TokenComment || (len(last.Bytes) > 0 && last.Bytes[len(last.Bytes)-1] != '\n') {
				body.AppendUnstructuredTokens(hclwrite.Tokens{
					&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
				})
			}
		}
	}
	return nil
}
