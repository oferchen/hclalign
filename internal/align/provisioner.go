// internal/align/provisioner.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type provisionerStrategy struct{}

func (provisionerStrategy) Name() string { return "provisioner" }

func (provisionerStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()

	ordered := []string{"when", "on_failure"}
	allowed := map[string]struct{}{ordered[0]: {}, ordered[1]: {}}

	type item struct {
		name  string
		attr  bool
		block *hclwrite.Block
	}

	tokens := body.BuildTokens(nil)
	depth := 0
	blocks := body.Blocks()
	blockMap := map[string][]*hclwrite.Block{}
	for _, b := range blocks {
		blockMap[b.Type()] = append(blockMap[b.Type()], b)
	}

	var items []item
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok.Type {
		case hclsyntax.TokenOBrace, hclsyntax.TokenOParen:
			depth++
		case hclsyntax.TokenCBrace, hclsyntax.TokenCParen:
			if depth > 0 {
				depth--
			}
		case hclsyntax.TokenIdent:
			if depth != 0 {
				continue
			}
			name := string(tok.Bytes)
			j := i + 1
			for j < len(tokens) && (tokens[j].Type == hclsyntax.TokenNewline || tokens[j].Type == hclsyntax.TokenComment) {
				j++
			}
			if j >= len(tokens) {
				continue
			}
			switch tokens[j].Type {
			case hclsyntax.TokenEqual:
				items = append(items, item{name: name, attr: true})
			case hclsyntax.TokenOBrace:
				if bs := blockMap[name]; len(bs) > 0 {
					items = append(items, item{name: name, block: bs[0]})
					blockMap[name] = bs[1:]
				}
			}
		}
	}

	attrTokensMap := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokensMap[name] = ihcl.ExtractAttrTokens(attr)
		body.RemoveAttribute(name)
	}
	for _, b := range blocks {
		body.RemoveBlock(b)
	}

	var order []item
	for _, k := range ordered {
		if _, ok := attrs[k]; ok {
			order = append(order, item{name: k, attr: true})
		}
	}
	for _, it := range items {
		if it.attr {
			if _, ok := allowed[it.name]; ok {
				continue
			}
			order = append(order, it)
			continue
		}
		order = append(order, it)
	}

	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	body.Clear()
	if len(order) > 0 {
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
	for _, it := range order {
		if it.attr {
			appendAttr(it.name)
			continue
		}
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
		body.AppendBlock(it.block)
	}
	if trailingComma && len(order) > 0 {
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

func init() { Register(provisionerStrategy{}) }
