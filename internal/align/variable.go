// /internal/align/variable.go
package align

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type variableStrategy struct{}

func (variableStrategy) Name() string { return "variable" }

func (variableStrategy) Align(block *hclwrite.Block, opts *Options) error {
	order := opts.Order
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
	if len(knownOrder) == 0 {
		knownOrder = config.CanonicalOrder
	}
	return reorderVariableBlock(block, knownOrder, canonicalSet, opts.Strict)
}

func init() {
	Register(variableStrategy{})
}

func reorderVariableBlock(block *hclwrite.Block, order []string, canonicalSet map[string]struct{}, strict bool) error {
	body := block.Body()

	attrs := body.Attributes()
	nestedBlocks := body.Blocks()

	if strict {
		var missing []string
		for name := range canonicalSet {
			if _, ok := attrs[name]; !ok {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			varName := ""
			if labels := block.Labels(); len(labels) > 0 {
				varName = labels[0]
			}
			return fmt.Errorf("variable %q: missing attributes: %s", varName, strings.Join(missing, ", "))
		}
		var unknown []string
		for name := range attrs {
			if _, ok := canonicalSet[name]; !ok {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			varName := ""
			if labels := block.Labels(); len(labels) > 0 {
				varName = labels[0]
			}
			return fmt.Errorf("variable %q: unknown attributes: %s", varName, strings.Join(unknown, ", "))
		}
	}

	allTokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(allTokens)
	prefixTokens := hclwrite.Tokens{}
	var tailTokens hclwrite.Tokens
	blockLeadTokens := make(map[*hclwrite.Block]hclwrite.Tokens)
	attrLeadTrim := make(map[string]int)
	attrExtraLead := make(map[string]hclwrite.Tokens)
	currentTokens := hclwrite.Tokens{}
	prefixCaptured := false
	blockIndex := 0
	capturedComments := 0
	for i := 0; i < len(allTokens); {
		tok := allTokens[i]
		if tok.Type == hclsyntax.TokenComment && !prefixCaptured {
			cpy := *tok
			if bytes.HasSuffix(cpy.Bytes, newline) {
				cpy.Bytes = cpy.Bytes[:len(cpy.Bytes)-len(newline)]
				prefixTokens = append(prefixTokens, &cpy)
				prefixTokens = append(prefixTokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline})
			} else {
				prefixTokens = append(prefixTokens, &cpy)
			}
			capturedComments++
			i++
			continue
		}
		if tok.Type == hclsyntax.TokenIdent {
			name := string(tok.Bytes)
			if attr, ok := attrs[name]; ok && i+1 < len(allTokens) && allTokens[i+1].Type == hclsyntax.TokenEqual {
				attrToks := attr.BuildTokens(nil)
				leadCount := 0
				for leadCount < len(attrToks) && attrToks[leadCount].Type == hclsyntax.TokenComment {
					leadCount++
				}
				if !prefixCaptured {
					prefixTokens = append(prefixTokens, currentTokens...)
					if capturedComments > 0 {
						attrLeadTrim[name] = capturedComments
						capturedComments = 0
					}
					prefixCaptured = true
				} else if len(currentTokens) > 0 && leadCount == 0 {
					attrExtraLead[name] = append(hclwrite.Tokens{}, currentTokens...)
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
					prefixTokens = append(prefixTokens, currentTokens...)
					capturedComments = 0
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
		prefixTokens = append(prefixTokens, currentTokens...)
		currentTokens = nil
	}
	tailTokens = currentTokens

	ihcl.NormalizeTokens(prefixTokens)
	for _, lead := range blockLeadTokens {
		ihcl.NormalizeTokens(lead)
	}
	ihcl.NormalizeTokens(tailTokens)

	for _, nb := range nestedBlocks {
		body.RemoveBlock(nb)
	}

	attrTokensMap := make(map[string]ihcl.AttrTokens)
	for name, attr := range attrs {
		at := ihcl.ExtractAttrTokens(attr)
		if extra, ok := attrExtraLead[name]; ok {
			at.LeadTokens = append(extra, at.LeadTokens...)
		}
		if trim := attrLeadTrim[name]; trim > 0 {
			if trim < len(at.LeadTokens) {
				at.LeadTokens = at.LeadTokens[trim:]
			} else {
				at.LeadTokens = nil
			}
		}
		attrTokensMap[name] = at
	}

	originalOrder := ihcl.AttributeOrder(body, attrs)

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
				body.AppendUnstructuredTokens(tok.LeadTokens)
				body.SetAttributeRaw(name, tok.ExprTokens)
			}
		}
		for _, name := range originalOrder {
			if _, isKnown := canonicalSet[name]; isKnown {
				continue
			}
			if tok, ok := attrTokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.LeadTokens)
				body.SetAttributeRaw(name, tok.ExprTokens)
			}
		}
	} else {
		finalOrder := make([]string, 0, len(originalOrder))

		finalOrder = append(finalOrder, orderedKnown...)
		for _, name := range originalOrder {
			if _, isKnown := canonicalSet[name]; !isKnown {
				finalOrder = append(finalOrder, name)
			}
		}

		for _, name := range finalOrder {
			if tok, ok := attrTokensMap[name]; ok {
				body.AppendUnstructuredTokens(tok.LeadTokens)
				body.SetAttributeRaw(name, tok.ExprTokens)
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
