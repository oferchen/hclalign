// internal/align/terraform.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type terraformStrategy struct{}

func (terraformStrategy) Name() string { return "terraform" }

func (terraformStrategy) Align(block *hclwrite.Block, _ *Options) error {
	body := block.Body()

	attrs := body.Attributes()
	blocks := body.Blocks()

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	order := ihcl.AttributeOrder(body, attrs)

	attrTokens := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokens[name] = ihcl.ExtractAttrTokens(attr, nil)
		body.RemoveAttribute(name)
	}

	canonical := CanonicalBlockAttrOrder["terraform"]
	canonicalSet := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		canonicalSet[name] = struct{}{}
	}

	blockByType := make(map[string]*hclwrite.Block, len(blocks))
	var otherBlocks []*hclwrite.Block
	for _, b := range blocks {
		body.RemoveBlock(b)
		if _, ok := canonicalSet[b.Type()]; ok {
			blockByType[b.Type()] = b
		} else {
			otherBlocks = append(otherBlocks, b)
		}
	}

	if rp := blockByType["required_providers"]; rp != nil {
		rpAttrs := rp.Body().Attributes()
		names := ihcl.AttributeOrder(rp.Body(), rpAttrs)
		if err := reorderBlock(rp, names); err != nil {
			return err
		}
	}

	canonicalAttrs := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		if _, ok := attrTokens[name]; ok {
			canonicalAttrs[name] = struct{}{}
		}
	}

	var otherAttrs []string
	for _, name := range order {
		if _, ok := canonicalAttrs[name]; ok {
			continue
		}
		otherAttrs = append(otherAttrs, name)
	}
	type item struct {
		name   string
		block  *hclwrite.Block
		isAttr bool
	}

	var items []item
	for _, name := range canonical {
		if _, ok := attrTokens[name]; ok {
			items = append(items, item{name: name, isAttr: true})
			continue
		}
		if b, ok := blockByType[name]; ok {
			items = append(items, item{block: b})
		}
	}
	for _, name := range otherAttrs {
		items = append(items, item{name: name, isAttr: true})
	}
	for _, b := range otherBlocks {
		items = append(items, item{block: b})
	}

	body.Clear()
	if len(items) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}
	for _, it := range items {
		if it.isAttr {
			tok := attrTokens[it.name]
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(it.name, tok.ExprTokens)
		} else {
			body.AppendUnstructuredTokens(hclwrite.Tokens{
				&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
			})
			body.AppendBlock(it.block)
		}
	}
	if trailingComma && len(items) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
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

func init() { Register(terraformStrategy{}) }
