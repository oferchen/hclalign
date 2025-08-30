// filename: internal/align/variable.go
package align

import (
	"bytes"
	"sort"

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
	validationPos := -1
	for _, name := range order {
		if name == "validation" {
			if validationPos == -1 {
				validationPos = len(knownOrder)
			}
			continue
		}
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
	return reorderVariableBlock(block, knownOrder, canonicalSet, validationPos, opts.PrefixOrder)
}

func init() {
	Register(variableStrategy{})
}

func reorderVariableBlock(block *hclwrite.Block, order []string, canonicalSet map[string]struct{}, validationPos int, prefixOrder bool) error {
	body := block.Body()

	attrs := body.Attributes()
	attrOrder := make([]string, 0, len(attrs))
	nestedBlocks := body.Blocks()

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
				attrOrder = append(attrOrder, name)
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
	validationBlocks := make([]*hclwrite.Block, 0)
	otherBlocks := make([]*hclwrite.Block, 0)
	for _, nb := range nestedBlocks {
		if nb.Type() == "validation" {
			validationBlocks = append(validationBlocks, nb)
		} else {
			otherBlocks = append(otherBlocks, nb)
		}
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

	for name := range attrs {
		body.RemoveAttribute(name)
	}

	body.Clear()
	body.AppendUnstructuredTokens(prefixTokens)

	canonicalOrderSet := map[string]struct{}{}
	orderedKnown := make([]string, 0, len(config.CanonicalOrder))
	for _, name := range order {
		if _, ok := canonicalSet[name]; !ok {
			continue
		}
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

	unknown := make([]string, 0)
	seenUnknown := make(map[string]struct{})
	for _, name := range attrOrder {
		if _, isKnown := canonicalSet[name]; !isKnown {
			if _, dup := seenUnknown[name]; dup {
				continue
			}
			unknown = append(unknown, name)
			seenUnknown[name] = struct{}{}
		}
	}
	if prefixOrder {
		sort.Strings(unknown)
	}

	insertedValidation := false
	for i, name := range orderedKnown {
		if !insertedValidation && validationPos != -1 && i == validationPos {
			for _, nb := range validationBlocks {
				if lead, ok := blockLeadTokens[nb]; ok {
					body.AppendUnstructuredTokens(lead)
				}
				body.AppendBlock(nb)
			}
			insertedValidation = true
		}
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}
	if !insertedValidation && validationPos != -1 {
		for _, nb := range validationBlocks {
			if lead, ok := blockLeadTokens[nb]; ok {
				body.AppendUnstructuredTokens(lead)
			}
			body.AppendBlock(nb)
		}
		insertedValidation = true
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

	for _, name := range unknown {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}

	if !insertedValidation {
		for _, nb := range validationBlocks {
			if lead, ok := blockLeadTokens[nb]; ok {
				body.AppendUnstructuredTokens(lead)
			}
			body.AppendBlock(nb)
		}
	}

	for _, nb := range otherBlocks {
		if lead, ok := blockLeadTokens[nb]; ok {
			body.AppendUnstructuredTokens(lead)
		}
		body.AppendBlock(nb)
	}
	body.AppendUnstructuredTokens(tailTokens)

	return nil
}
