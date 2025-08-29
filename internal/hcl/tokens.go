package hcl

import (
	"bytes"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// AttrTokens splits an attribute's tokens into leading comment tokens and
// expression tokens. The resulting slices omit trailing newlines or comments
// that include a newline so they can be reinserted with SetAttributeRaw.
type AttrTokens struct {
	LeadTokens hclwrite.Tokens
	ExprTokens hclwrite.Tokens
}

// ExtractAttrTokens returns the leading comment tokens and expression tokens
// for the given attribute. The returned slices are suitable for reinsertion via
// SetAttributeRaw.
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

// HasTrailingComma reports whether the body tokens contain a trailing comma
// immediately before the closing brace.
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

// DetectLineEnding inspects tokens and returns the newline sequence used.
// It defaults to LF if no newline token is found.
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

// NormalizeTokens converts CRLF line endings to LF in the provided tokens and
// strips any UTF-8 BOM from the first token. The returned boolean indicates
// whether a BOM was removed.
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

// AttributeOrder returns the order of attributes as they originally appear in
// the body. The attrs map should contain the attributes present in the body.
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
