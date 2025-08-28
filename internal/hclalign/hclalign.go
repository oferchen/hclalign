// internal/hclalign/hclalign.go
package hclalign

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
)

func ReorderAttributes(file *hclwrite.File, order []string, strict bool) error {
	if len(order) == 0 {
		order = config.CanonicalOrder
	}

	canonicalSet := make(map[string]struct{}, len(config.CanonicalOrder))
	for _, name := range config.CanonicalOrder {
		canonicalSet[name] = struct{}{}
	}

	knownOrder := make([]string, 0, len(order))
	seen := make(map[string]struct{}, len(order))
	for _, name := range order {
		if _, ok := canonicalSet[name]; ok {
			if _, dup := seen[name]; dup {
				continue
			}
			knownOrder = append(knownOrder, name)
			seen[name] = struct{}{}
		}
	}

	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() != "variable" {
			continue
		}

		if err := reorderVariableBlock(block, knownOrder, canonicalSet, strict); err != nil {
			return err
		}
	}

	return nil
}

func reorderVariableBlock(block *hclwrite.Block, order []string, canonicalSet map[string]struct{}, strict bool) error {
	body := block.Body()

	attrs := body.Attributes()
	nestedBlocks := body.Blocks()

	if strict {
		var unknown, missing []string
		for name := range attrs {
			if _, ok := canonicalSet[name]; !ok {
				unknown = append(unknown, name)
			}
		}
		for name := range canonicalSet {
			if _, ok := attrs[name]; !ok {
				missing = append(missing, name)
			}
		}
		if len(unknown) > 0 || len(missing) > 0 {
			sort.Strings(unknown)
			sort.Strings(missing)
			var parts []string
			if len(unknown) > 0 {
				parts = append(parts, fmt.Sprintf("unknown attributes: %s", strings.Join(unknown, ", ")))
			}
			if len(missing) > 0 {
				parts = append(parts, fmt.Sprintf("missing attributes: %s", strings.Join(missing, ", ")))
			}
			varName := ""
			if labels := block.Labels(); len(labels) > 0 {
				varName = labels[0]
			}
			return fmt.Errorf("variable %q: %s", varName, strings.Join(parts, "; "))
		}
	}

	allTokens := body.BuildTokens(nil)
	prefixTokens := hclwrite.Tokens{}
	tailTokens := hclwrite.Tokens{}
	blockLeadTokens := make(map[*hclwrite.Block]hclwrite.Tokens)
	currentTokens := hclwrite.Tokens{}
	prefixCaptured := false
	blockIndex := 0
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
				if !prefixCaptured {
					if len(currentTokens) >= leadCount {
						prefixTokens = append(hclwrite.Tokens{}, currentTokens[:len(currentTokens)-leadCount]...)
					}
					prefixCaptured = true
				}
				currentTokens = nil
				i += len(attrToks) - leadCount
				continue
			}
			if blockIndex < len(nestedBlocks) {
				blockToks := nestedBlocks[blockIndex].BuildTokens(nil)
				leadCount := 0
				for leadCount < len(blockToks) && blockToks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !prefixCaptured {
					if len(currentTokens) >= leadCount {
						prefixTokens = append(hclwrite.Tokens{}, currentTokens[:len(currentTokens)-leadCount]...)
					}
					prefixCaptured = true
				} else {
					if len(currentTokens) >= leadCount {
						blockLeadTokens[nestedBlocks[blockIndex]] = append(hclwrite.Tokens{}, currentTokens[:len(currentTokens)-leadCount]...)
					}
				}
				currentTokens = nil
				i += len(blockToks) - leadCount
				blockIndex++
				continue
			}
		}
		currentTokens = append(currentTokens, tok)
		i++
	}
	if !prefixCaptured {
		prefixTokens = currentTokens
		currentTokens = nil
	}
	tailTokens = currentTokens

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
	normalizeTokens(prefixTokens)
	for _, lead := range blockLeadTokens {
		normalizeTokens(lead)
	}
	normalizeTokens(tailTokens)

	for _, nb := range nestedBlocks {
		body.RemoveBlock(nb)
	}

	attrTokensMap := make(map[string]attrTokens)
	for name, attr := range attrs {
		attrTokensMap[name] = extractAttrTokens(attr)
	}

	originalOrder := attributeOrder(body, attrs)

	for name := range attrs {
		body.RemoveAttribute(name)
	}

	body.Clear()
	body.AppendUnstructuredTokens(prefixTokens)

	canonicalOrderSet := map[string]struct{}{}
	orderedKnown := make([]string, 0, len(config.CanonicalOrder))
	for _, name := range order {
		canonicalOrderSet[name] = struct{}{}
		if _, ok := attrTokensMap[name]; ok {
			orderedKnown = append(orderedKnown, name)
		}
	}

	for _, name := range config.CanonicalOrder {
		if _, already := canonicalOrderSet[name]; already {
			continue
		}
		if _, ok := attrTokensMap[name]; ok {
			orderedKnown = append(orderedKnown, name)
		}
	}

	if strict {

		for _, name := range orderedKnown {
			if tok, ok := attrTokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.leadTokens)
				body.SetAttributeRaw(name, tok.exprTokens)
			}
		}
		for _, name := range originalOrder {
			if _, isKnown := canonicalSet[name]; isKnown {
				continue
			}
			if tok, ok := attrTokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.leadTokens)
				body.SetAttributeRaw(name, tok.exprTokens)
			}
		}
	} else {

		idx := 0
		for _, name := range originalOrder {
			if _, isKnown := canonicalSet[name]; isKnown {
				if idx < len(orderedKnown) {
					k := orderedKnown[idx]
					idx++
					if tok, ok := attrTokensMap[k]; ok {
						body.AppendUnstructuredTokens(tok.leadTokens)
						body.SetAttributeRaw(k, tok.exprTokens)
					}
				}
				continue
			}
			if tok, ok := attrTokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.leadTokens)
				body.SetAttributeRaw(name, tok.exprTokens)
			}
		}
	}

	hasLeadingNewline := false
	for _, t := range prefixTokens {
		if t.Type == hclsyntax.TokenNewline {
			hasLeadingNewline = true
			break
		}
	}
	if !hasLeadingNewline && len(nestedBlocks) == 0 && (len(tailTokens) == 0 || tailTokens[0].Type != hclsyntax.TokenNewline) {
		toks := body.BuildTokens(nil)
		if n := len(toks); n > 0 && toks[n-1].Type == hclsyntax.TokenNewline {
			body.Clear()
			body.AppendUnstructuredTokens(toks[:n-1])
		}
	}

	for _, nb := range nestedBlocks {
		if lead, ok := blockLeadTokens[nb]; ok {
			body.AppendUnstructuredTokens(lead)
		}
		body.AppendBlock(nb)
	}
	body.AppendUnstructuredTokens(tailTokens)

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
