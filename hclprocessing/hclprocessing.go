// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var canonicalOrder = []string{"description", "type", "default", "sensitive", "nullable"}

// ReorderAttributes reorders attributes of "variable" blocks into the provided
// order. When strict is true, unknown attributes are placed after known ones.
// Otherwise, unknown attributes remain in their original positions. If order is
// nil or empty, a canonical order is used.
func ReorderAttributes(file *hclwrite.File, order []string, strict bool) {
	if len(order) == 0 {
		order = canonicalOrder
	}

	// knownSet contains the canonical attributes for quick lookups.
	knownSet := make(map[string]struct{}, len(canonicalOrder))
	for _, name := range canonicalOrder {
		knownSet[name] = struct{}{}
	}

	// Filter the provided order to only known attributes so that unknown
	// attributes are appended after all known attributes regardless of their
	// position in the provided order.
	filtered := make([]string, 0, len(order))
	seen := make(map[string]struct{}, len(order))
	for _, name := range order {
		if _, ok := knownSet[name]; ok {
			if _, dup := seen[name]; dup {
				continue
			}
			filtered = append(filtered, name)
			seen[name] = struct{}{}
		}
	}

	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() != "variable" {
			continue
		}

		reorderVariableBlock(block, filtered, knownSet, strict)
	}
}

func reorderVariableBlock(block *hclwrite.Block, order []string, knownSet map[string]struct{}, strict bool) {
	body := block.Body()

	// Preserve nested blocks to re-append later.
	nested := body.Blocks()
	for _, nb := range nested {
		body.RemoveBlock(nb)
	}

	attrs := body.Attributes()
	tokensMap := make(map[string]attrTokens)
	for name, attr := range attrs {
		tokensMap[name] = extractAttrTokens(attr)
	}

	// Capture original attribute ordering.
	orderedNames := attributeOrder(body, attrs)

	for name := range attrs {
		body.RemoveAttribute(name)
	}

	tail := body.BuildTokens(nil)
	hasNewline := len(tail) > 0 && tail[0].Type == hclsyntax.TokenNewline
	if hasNewline {
		tail = tail[1:]
	}
	body.Clear()
	if hasNewline {
		body.AppendNewline()
	}

	canonSet := map[string]struct{}{}
	finalKnown := make([]string, 0, len(canonicalOrder))
	for _, name := range order {
		canonSet[name] = struct{}{}
		if _, ok := tokensMap[name]; ok {
			finalKnown = append(finalKnown, name)
		}
	}

	// Append remaining known attributes not specified in order following the
	// canonical order.
	for _, name := range canonicalOrder {
		if _, already := canonSet[name]; already {
			continue
		}
		if _, ok := tokensMap[name]; ok {
			finalKnown = append(finalKnown, name)
		}
	}

	if strict {
		// Place known attributes first followed by unknown attributes.
		for _, name := range finalKnown {
			if tok, ok := tokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.lead)
				body.SetAttributeRaw(name, tok.expr)
			}
		}
		for _, name := range orderedNames {
			if _, isKnown := knownSet[name]; isKnown {
				continue
			}
			if tok, ok := tokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.lead)
				body.SetAttributeRaw(name, tok.expr)
			}
		}
	} else {
		// Merge known attributes in canonical order with unknown
		// attributes at their original positions.
		idx := 0
		for _, name := range orderedNames {
			if _, isKnown := knownSet[name]; isKnown {
				if idx < len(finalKnown) {
					k := finalKnown[idx]
					idx++
					if tok, ok := tokensMap[k]; ok {
						body.AppendUnstructuredTokens(tok.lead)
						body.SetAttributeRaw(k, tok.expr)
					}
				}
				continue
			}
			if tok, ok := tokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.lead)
				body.SetAttributeRaw(name, tok.expr)
			}
		}
	}

	if !hasNewline {
		toks := body.BuildTokens(nil)
		if n := len(toks); n > 0 && toks[n-1].Type == hclsyntax.TokenNewline {
			body.Clear()
			body.AppendUnstructuredTokens(toks[:n-1])
		}
	}
	body.AppendUnstructuredTokens(tail)

	for _, nb := range nested {
		body.AppendBlock(nb)
	}
}

type attrTokens struct {
	lead hclwrite.Tokens
	expr hclwrite.Tokens
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
	return attrTokens{lead: lead, expr: expr}
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
