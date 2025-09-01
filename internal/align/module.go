// internal/align/module.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type moduleStrategy struct{}

func (moduleStrategy) Name() string { return "module" }

func (moduleStrategy) Align(block *hclwrite.Block, _ *Options) error {
	body := block.Body()
	attrs := body.Attributes()
	canonical := CanonicalBlockAttrOrder["module"]

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	attrTokens := make(map[string]ihcl.AttrTokens, len(attrs))
	for name, attr := range attrs {
		attrTokens[name] = ihcl.ExtractAttrTokens(attr, nil)
	}
	blocks := body.Blocks()

	type segment struct {
		attrs []string
		block *hclwrite.Block
	}
	segments := []segment{}
	current := []string{}
	depth := 0
	bi := 0
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok.Type {
		case hclsyntax.TokenOBrace, hclsyntax.TokenOParen:
			depth++
			continue
		case hclsyntax.TokenCBrace, hclsyntax.TokenCParen:
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth == 0 && tok.Type == hclsyntax.TokenIdent {
			name := string(tok.Bytes)
			if _, ok := attrs[name]; ok && tokens[i+1].Type == hclsyntax.TokenEqual {
				current = append(current, name)
				continue
			}
			if bi < len(blocks) {
				if len(current) > 0 {
					segments = append(segments, segment{attrs: current})
					current = nil
				}
				segments = append(segments, segment{block: blocks[bi]})
				btoks := blocks[bi].BuildTokens(nil)
				bi++
				i += len(btoks) - 1
			}
		}
	}
	if len(current) > 0 {
		segments = append(segments, segment{attrs: current})
	}

	for name := range attrs {
		body.RemoveAttribute(name)
	}
	for _, b := range blocks {
		body.RemoveBlock(b)
	}
	body.Clear()

	if len(segments) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}
	for i, seg := range segments {
		if i > 0 {
			body.AppendUnstructuredTokens(hclwrite.Tokens{
				&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
			})
		}
		if seg.block != nil {
			body.AppendBlock(seg.block)
			continue
		}
		ordered := orderModuleAttrs(seg.attrs, canonical)
		for _, name := range ordered {
			tok := attrTokens[name]
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}

	if trailingComma && len(segments) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}},
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

func orderModuleAttrs(names, canonical []string) []string {
	canonicalSet := make(map[string]struct{}, len(canonical))
	for _, c := range canonical {
		canonicalSet[c] = struct{}{}
	}
	order := make([]string, 0, len(names))
	for _, c := range canonical {
		for _, n := range names {
			if n == c {
				order = append(order, n)
			}
		}
	}
	for _, n := range names {
		if _, ok := canonicalSet[n]; !ok {
			order = append(order, n)
		}
	}
	return order
}

func init() { Register(moduleStrategy{}) }
