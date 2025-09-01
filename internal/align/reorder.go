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

	attrPre := map[string]hclwrite.Tokens{}
	blockPre := make([]hclwrite.Tokens, len(nestedBlocks))
	prefixTokens := hclwrite.Tokens{}
	current := hclwrite.Tokens{}
	prefixCaptured := false
	blockIndex := 0
	for i := 0; i < len(tokens); {
		tok := tokens[i]
		if tok.Type == hclsyntax.TokenIdent {
			name := string(tok.Bytes)
			if attr, ok := attrs[name]; ok && i+1 < len(tokens) && tokens[i+1].Type == hclsyntax.TokenEqual {
				attrToks := attr.BuildTokens(nil)
				leadCount := 0
				for leadCount < len(attrToks) && attrToks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !prefixCaptured {
					prefixTokens = append(prefixTokens, current...)
					prefixCaptured = true
				} else if len(current) > leadCount {
					attrPre[name] = append(hclwrite.Tokens{}, current[:len(current)-leadCount]...)
				}
				current = nil
				i += len(attrToks) - leadCount
				continue
			}
			if blockIndex < len(nestedBlocks) {
				btoks := nestedBlocks[blockIndex].BuildTokens(nil)
				leadCount := 0
				for leadCount < len(btoks) && btoks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !prefixCaptured {
					prefixTokens = append(prefixTokens, current...)
					prefixCaptured = true
				} else if len(current) > leadCount {
					blockPre[blockIndex] = append(hclwrite.Tokens{}, current[:len(current)-leadCount]...)
				}
				current = nil
				i += len(btoks) - leadCount
				blockIndex++
				continue
			}
		}
		current = append(current, tok)
		i++
	}
	if !prefixCaptured {
		prefixTokens = append(prefixTokens, current...)
	}

	attrTokensMap := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokensMap[name] = ihcl.ExtractAttrTokens(attr, attrPre[name])
	}

	for name := range attrs {
		body.RemoveAttribute(name)
	}
	for _, nb := range nestedBlocks {
		body.RemoveBlock(nb)
	}

	body.Clear()
	body.AppendUnstructuredTokens(prefixTokens)
	for _, name := range order {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.PreTokens)
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
			body.AppendUnstructuredTokens(tok.TrailTokens)
		}
	}
	for i, nb := range nestedBlocks {
		if pre := blockPre[i]; len(pre) > 0 {
			body.AppendUnstructuredTokens(pre)
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
