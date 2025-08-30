// internal/align/module.go
package align

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type moduleStrategy struct{}

func (moduleStrategy) Name() string { return "module" }

func (moduleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	if opts != nil && opts.PrefixOrder {
		if attr, ok := attrs["providers"]; ok {
			toks := attr.Expr().BuildTokens(nil)
			if len(toks) > 0 && toks[0].Type == hclsyntax.TokenOBrace {
				sortProviders(block.Body(), attr)
				attrs = block.Body().Attributes()
			}
		}
	}

	canonical := CanonicalBlockAttrOrder["module"]
	order := make([]string, 0, len(attrs))
	reserved := make(map[string]struct{}, len(canonical))
	for _, name := range canonical {
		if _, ok := attrs[name]; ok {
			order = append(order, name)
		}
		reserved[name] = struct{}{}
	}

	vars := make([]string, 0, len(attrs))
	for name := range attrs {
		if _, ok := reserved[name]; !ok {
			vars = append(vars, name)
		}
	}
	sort.Strings(vars)

	order = append(order, vars...)

	return reorderBlock(block, order)
}

func sortProviders(body *hclwrite.Body, attr *hclwrite.Attribute) {
	tokens := attr.Expr().BuildTokens(nil)
	if len(tokens) < 3 || tokens[0].Type != hclsyntax.TokenOBrace {
		return
	}
	prefix := hclwrite.Tokens{}
	i := 1
	for i < len(tokens)-1 && tokens[i].Type != hclsyntax.TokenIdent {
		prefix = append(prefix, tokens[i])
		i++
	}
	start := i
	type kv struct {
		name   string
		tokens hclwrite.Tokens
	}
	var kvs []kv
	for start < len(tokens)-1 {
		if tokens[start].Type != hclsyntax.TokenIdent {
			break
		}
		name := string(tokens[start].Bytes)
		i := start + 1
		depth := 0
		for i < len(tokens)-1 {
			t := tokens[i]
			switch t.Type {
			case hclsyntax.TokenOBrace, hclsyntax.TokenOBrack, hclsyntax.TokenOParen:
				depth++
			case hclsyntax.TokenCBrace, hclsyntax.TokenCBrack, hclsyntax.TokenCParen:
				if depth > 0 {
					depth--
				}
			case hclsyntax.TokenComma:
				if depth == 0 {
					i++
					if i < len(tokens)-1 && tokens[i].Type == hclsyntax.TokenNewline {
						i++
					}
					goto set
				}
			case hclsyntax.TokenNewline:
				if depth == 0 {
					i++
					goto set
				}
			}
			i++
		}
	set:
		kvs = append(kvs, kv{name: name, tokens: append(hclwrite.Tokens{}, tokens[start:i]...)})
		start = i
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].name < kvs[j].name })
	expr := hclwrite.Tokens{tokens[0]}
	expr = append(expr, prefix...)
	for _, kv := range kvs {
		expr = append(expr, kv.tokens...)
	}
	expr = append(expr, tokens[len(tokens)-1])
	body.SetAttributeRaw("providers", expr)
}

func init() { Register(moduleStrategy{}) }
