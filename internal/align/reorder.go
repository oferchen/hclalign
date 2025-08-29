package align

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// reorderBlock reorders attributes in the given block according to order.
// The order slice should contain all attribute names in their desired final
// positions. Tokens and comments are preserved using BuildTokens and
// SetAttributeRaw.
func reorderBlock(block *hclwrite.Block, order []string) error {
	body := block.Body()
	attrs := body.Attributes()
	nestedBlocks := body.Blocks()

	attrTokensMap := map[string]attrTokens{}
	for name, attr := range attrs {
		attrTokensMap[name] = extractAttrTokens(attr)
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
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		})
	}
	for _, name := range order {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.leadTokens)
			body.SetAttributeRaw(name, tok.exprTokens)
		}
	}
	for _, nb := range nestedBlocks {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		})
		body.AppendBlock(nb)
	}
	toks := body.BuildTokens(nil)
	if len(toks) > 0 && toks[len(toks)-1].Type != hclsyntax.TokenNewline {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		})
	}
	return nil
}

type attrTokens struct {
	leadTokens hclwrite.Tokens
	exprTokens hclwrite.Tokens
}

func extractAttrTokens(attr *hclwrite.Attribute) attrTokens {
	toks := attr.BuildTokens(nil)
	i := 0
	for i < len(toks) && toks[i].Type == hclsyntax.TokenComment {
		i++
	}
	lead := toks[:i]
	expr := toks[i+2:]
	if n := len(expr); n > 0 {
		last := expr[n-1]
		if last.Type == hclsyntax.TokenNewline {
			expr = expr[:n-1]
		} else if last.Type == hclsyntax.TokenComment {
			b := last.Bytes
			if len(b) > 0 && b[len(b)-1] == '\n' {
				expr[n-1].Bytes = b[:len(b)-1]
			}
		}
	}
	return attrTokens{leadTokens: lead, exprTokens: expr}
}

func attributeOrder(body *hclwrite.Body, attrs map[string]*hclwrite.Attribute) []string {
	tokens := body.BuildTokens(nil)
	order := make([]string, 0, len(attrs))
	depth := 0
	for i := 0; i < len(tokens)-1; i++ {
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
				order = append(order, name)
			}
		}
	}
	return order
}
