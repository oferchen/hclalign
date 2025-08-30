// internal/hcl/tokens.go
package hcl

import (
	"bytes"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type AttrTokens struct {
	LeadTokens hclwrite.Tokens
	ExprTokens hclwrite.Tokens
}

func ExtractAttrTokens(attr *hclwrite.Attribute) AttrTokens {
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
	return AttrTokens{LeadTokens: lead, ExprTokens: expr}
}

func HasTrailingComma(tokens hclwrite.Tokens) bool {
	for i := len(tokens) - 1; i >= 0; i-- {
		tok := tokens[i]
		if tok.Type == hclsyntax.TokenCBrace {
			for j := i - 1; j >= 0; j-- {
				t := tokens[j]
				switch t.Type {
				case hclsyntax.TokenNewline, hclsyntax.TokenComment:
					continue
				case hclsyntax.TokenComma:
					return true
				default:
					return false
				}
			}
		}
	}
	return false
}

func DetectLineEnding(tokens hclwrite.Tokens) []byte {
	for _, t := range tokens {
		if i := bytes.IndexByte(t.Bytes, '\n'); i >= 0 {
			if i > 0 && t.Bytes[i-1] == '\r' {
				return []byte{'\r', '\n'}
			}
			return []byte{'\n'}
		}
	}
	return []byte{'\n'}
}

func NormalizeTokens(tokens hclwrite.Tokens) bool {
	bom := []byte{0xEF, 0xBB, 0xBF}
	hasBOM := false
	for i, t := range tokens {
		b := t.Bytes
		if i == 0 && bytes.HasPrefix(b, bom) {
			hasBOM = true
			b = b[len(bom):]
		}
		if bytes.Contains(b, []byte{'\r'}) {
			b = bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
			b = bytes.ReplaceAll(b, []byte{'\r'}, nil)
		}
		t.Bytes = b
	}
	return hasBOM
}

func AttributeOrder(body *hclwrite.Body, attrs map[string]*hclwrite.Attribute) []string {
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
