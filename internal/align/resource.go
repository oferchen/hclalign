// /internal/align/resource.go
package align

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type resourceStrategy struct{}

func (resourceStrategy) Name() string { return "resource" }

func (resourceStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return schemaAwareOrder(block, opts)
}

func init() { Register(resourceStrategy{}) }

func schemaAwareOrder(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()
	originalOrder := ihcl.AttributeOrder(body, attrs)
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	if opts == nil || opts.Schema == nil {
		metaAttrs := []string{}
		for _, n := range []string{"provider", "count", "for_each", "depends_on"} {
			if _, ok := attrs[n]; ok {
				metaAttrs = append(metaAttrs, n)
			}
		}
		metaSet := map[string]struct{}{}
		for _, n := range metaAttrs {
			metaSet[n] = struct{}{}
		}
		var rest []string
		for _, n := range originalOrder {
			if _, ok := metaSet[n]; !ok {
				rest = append(rest, n)
			}
		}
		if opts.SortUnknown {
			sort.Strings(rest)
		}
		order := append(metaAttrs, rest...)
		return reorderBlock(block, order)
	}

	var req, opt, comp []string
	for _, name := range names {
		if _, ok := opts.Schema.Required[name]; ok {
			req = append(req, name)
			continue
		}
		if _, ok := opts.Schema.Optional[name]; ok {
			opt = append(opt, name)
			continue
		}
		if _, ok := opts.Schema.Computed[name]; ok {
			comp = append(comp, name)
			continue
		}
	}
	sort.Strings(req)
	sort.Strings(opt)
	sort.Strings(comp)

	metaAttrs := []string{}
	for _, n := range []string{"provider", "count", "for_each", "depends_on"} {
		if _, ok := attrs[n]; ok {
			metaAttrs = append(metaAttrs, n)
		}
	}

	known := map[string]struct{}{}
	for _, n := range metaAttrs {
		known[n] = struct{}{}
	}
	for _, n := range req {
		known[n] = struct{}{}
	}
	for _, n := range opt {
		known[n] = struct{}{}
	}
	for _, n := range comp {
		known[n] = struct{}{}
	}

	var unk []string
	for _, n := range originalOrder {
		if _, ok := known[n]; !ok {
			unk = append(unk, n)
		}
	}
	if opts.SortUnknown {
		sort.Strings(unk)
	}

	if opts.Strict {
		for r := range opts.Schema.Required {
			if _, ok := attrs[r]; !ok {
				return fmt.Errorf("missing required attribute %q", r)
			}
		}
	}

	blocks := body.Blocks()
	var lifecycleBlock *hclwrite.Block
	var provisionerBlocks []*hclwrite.Block
	var otherBlocks []*hclwrite.Block
	for _, nb := range blocks {
		switch nb.Type() {
		case "lifecycle":
			lifecycleBlock = nb
		case "provisioner":
			provisionerBlocks = append(provisionerBlocks, nb)
		default:
			otherBlocks = append(otherBlocks, nb)
		}
	}

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	attrTokensMap := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokensMap[name] = ihcl.ExtractAttrTokens(attr)
		body.RemoveAttribute(name)
	}
	for _, nb := range blocks {
		body.RemoveBlock(nb)
	}

	body.Clear()
	if len(metaAttrs) > 0 || lifecycleBlock != nil || len(provisionerBlocks) > 0 || len(req) > 0 || len(opt) > 0 || len(comp) > 0 || len(unk) > 0 || len(otherBlocks) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}

	appendAttr := func(name string) {
		if tok, ok := attrTokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}

	for _, name := range metaAttrs {
		appendAttr(name)
	}

	if lifecycleBlock != nil {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
		body.AppendBlock(lifecycleBlock)
	}
	for _, nb := range provisionerBlocks {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
		body.AppendBlock(nb)
	}

	for _, name := range req {
		appendAttr(name)
	}
	for _, name := range opt {
		appendAttr(name)
	}
	for _, name := range comp {
		appendAttr(name)
	}
	for _, name := range unk {
		appendAttr(name)
	}

	for _, nb := range otherBlocks {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
		body.AppendBlock(nb)
	}

	if trailingComma && (len(metaAttrs) > 0 || lifecycleBlock != nil || len(provisionerBlocks) > 0 || len(req) > 0 || len(opt) > 0 || len(comp) > 0 || len(unk) > 0 || len(otherBlocks) > 0) {
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
