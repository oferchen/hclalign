// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"bytes"

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

	attrs := body.Attributes()
	nested := body.Blocks()

	// Capture prefix tokens before any attributes or blocks, the tail tokens after
	// the last element, and any leading tokens for nested blocks so that we can
	// restore them later.
	allTokens := body.BuildTokens(nil)
	prefix := hclwrite.Tokens{}
	tail := hclwrite.Tokens{}
	blockLeads := make(map[*hclwrite.Block]hclwrite.Tokens)
	cur := hclwrite.Tokens{}
	seen := false
	bi := 0
	for i := 0; i < len(allTokens); {
		tok := allTokens[i]
		if tok.Type == hclsyntax.TokenIdent {
			name := string(tok.Bytes)
			if attr, ok := attrs[name]; ok && i+1 < len(allTokens) && allTokens[i+1].Type == hclsyntax.TokenEqual {
				attrToks := attr.BuildTokens(nil)
				leadCount := 0
				for leadCount < len(attrToks) && attrToks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !seen {
					if len(cur) >= leadCount {
						prefix = append(hclwrite.Tokens{}, cur[:len(cur)-leadCount]...)
					}
					seen = true
				}
				cur = nil
				i += len(attrToks) - leadCount
				continue
			}
			if bi < len(nested) {
				blockToks := nested[bi].BuildTokens(nil)
				leadCount := 0
				for leadCount < len(blockToks) && blockToks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !seen {
					if len(cur) >= leadCount {
						prefix = append(hclwrite.Tokens{}, cur[:len(cur)-leadCount]...)
					}
					seen = true
				} else {
					if len(cur) >= leadCount {
						blockLeads[nested[bi]] = append(hclwrite.Tokens{}, cur[:len(cur)-leadCount]...)
					}
				}
				cur = nil
				i += len(blockToks) - leadCount
				bi++
				continue
			}
		}
		cur = append(cur, tok)
		i++
	}
	if !seen {
		prefix = cur
		cur = nil
	}
	tail = cur

	normalizeTokens := func(toks hclwrite.Tokens) {
		for _, t := range toks {
			b := t.Bytes
			if bytes.Contains(b, []byte{'\r'}) {
				b = bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
				b = bytes.ReplaceAll(b, []byte{'\r'}, nil)
				t.Bytes = b
			}
		}
	}
	normalizeTokens(prefix)
	for _, lead := range blockLeads {
		normalizeTokens(lead)
	}
	normalizeTokens(tail)

	// Preserve nested blocks to re-append later.
	for _, nb := range nested {
		body.RemoveBlock(nb)
	}

	tokensMap := make(map[string]attrTokens)
	for name, attr := range attrs {
		tokensMap[name] = extractAttrTokens(attr)
	}

	// Capture original attribute ordering.
	orderedNames := attributeOrder(body, attrs)

	for name := range attrs {
		body.RemoveAttribute(name)
	}

	body.Clear()
	body.AppendUnstructuredTokens(prefix)

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

	hasLeadingNewline := false
	for _, t := range prefix {
		if t.Type == hclsyntax.TokenNewline {
			hasLeadingNewline = true
			break
		}
	}
	if !hasLeadingNewline && len(nested) == 0 && (len(tail) == 0 || tail[0].Type != hclsyntax.TokenNewline) {
		toks := body.BuildTokens(nil)
		if n := len(toks); n > 0 && toks[n-1].Type == hclsyntax.TokenNewline {
			body.Clear()
			body.AppendUnstructuredTokens(toks[:n-1])
		}
	}

	for _, nb := range nested {
		if lead, ok := blockLeads[nb]; ok {
			body.AppendUnstructuredTokens(lead)
		}
		body.AppendBlock(nb)
	}
	body.AppendUnstructuredTokens(tail)
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
	var order []string
	for i := 0; i < len(tokens)-1; i++ {
		tok := tokens[i]
		if tok.Type == hclsyntax.TokenIdent {
			name := string(tok.Bytes)
			if _, ok := attrs[name]; ok && tokens[i+1].Type == hclsyntax.TokenEqual {
				order = append(order, name)
			}
		}
	}
	return order
}
