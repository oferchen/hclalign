// internal/align/module.go
package align

import (
	"bytes"
	"sort"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type moduleStrategy struct{}

func (moduleStrategy) Name() string { return "module" }

func (moduleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()
	attrs := body.Attributes()
	canonical := CanonicalBlockAttrOrder["module"]

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	attrTokens := make(map[string]ihcl.AttrTokens, len(attrs))
	for name, attr := range attrs {
		attrTokens[name] = ihcl.ExtractAttrTokens(attr)
	}
	if opts != nil && opts.PrefixOrder {
		if at, ok := attrTokens["providers"]; ok {
			sorted, err := sortProvidersMap(at)
			if err != nil {
				return err
			}
			attrTokens["providers"] = sorted
		}
	}
	blocks := body.Blocks()

	type segment struct {
		attrs []string
		block *hclwrite.Block
	}
	segments := []segment{}
	current := []string{}
	depth := 0
	bi := 0
	for i := 0; i < len(tokens); i++ {
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
				current = append(current, name)
				continue
			}
			if bi < len(blocks) {
				if len(current) > 0 {
					segments = append(segments, segment{attrs: current})
					current = nil
				}
				segments = append(segments, segment{block: blocks[bi]})
				btoks := blocks[bi].BuildTokens(nil)
				bi++
				i += len(btoks) - 1
			}
		}
	}
	if len(current) > 0 {
		segments = append(segments, segment{attrs: current})
	}

	for name := range attrs {
		body.RemoveAttribute(name)
	}
	for _, b := range blocks {
		body.RemoveBlock(b)
	}
	body.Clear()

	if len(segments) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
		})
	}
	for i, seg := range segments {
		if i > 0 {
			body.AppendUnstructuredTokens(hclwrite.Tokens{
				&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: newline},
			})
		}
		if seg.block != nil {
			body.AppendBlock(seg.block)
			continue
		}
		ordered := orderModuleAttrs(seg.attrs, canonical)
		for _, name := range ordered {
			tok := attrTokens[name]
			if name == "providers" && opts != nil && opts.PrefixOrder {
				tok.ExprTokens = sortProviders(tok.ExprTokens)
			}
			body.AppendUnstructuredTokens(tok.LeadTokens)
			body.SetAttributeRaw(name, tok.ExprTokens)
		}
	}

	if trailingComma && len(segments) > 0 {
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte{','}},
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

func orderModuleAttrs(names, canonical []string) []string {
	reserved := make(map[string]struct{}, len(canonical))
	for _, n := range canonical {
		reserved[n] = struct{}{}
	}
	order := make([]string, 0, len(names))
	for _, n := range canonical {
		for _, m := range names {
			if m == n {
				order = append(order, m)
			}
		}
	}
	vars := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := reserved[n]; !ok {
			vars = append(vars, n)
		}
	}
	sort.Strings(vars)
	order = append(order, vars...)
	return order
}

func sortProviders(tokens hclwrite.Tokens) hclwrite.Tokens {
	buf := make([]byte, 0, len(tokens))
	for _, t := range tokens {
		buf = append(buf, t.Bytes...)
	}
	s := strings.TrimSpace(string(buf))
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return tokens
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	if body == "" {
		return tokens
	}
	lines := strings.Split(body, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	sort.Strings(items)
	var b strings.Builder
	b.WriteString("providers = {\n")
	for _, it := range items {
		b.WriteString("  ")
		b.WriteString(it)
		b.WriteByte('\n')
	}
	b.WriteString("}\n")
	f, diags := hclwrite.ParseConfig([]byte(b.String()), "p.tf", hcl.InitialPos)
	if diags.HasErrors() {
		return tokens
	}
	return ihcl.ExtractAttrTokens(f.Body().GetAttribute("providers")).ExprTokens
}

func init() { Register(moduleStrategy{}) }

func sortProvidersMap(at ihcl.AttrTokens) (ihcl.AttrTokens, error) {
	exprTokens := at.ExprTokens
	if len(exprTokens) < 2 || exprTokens[0].Type != hclsyntax.TokenOBrace {
		return at, nil
	}
	var buf bytes.Buffer
	buf.WriteString("dummy ")
	for _, t := range exprTokens {
		buf.Write(t.Bytes)
	}
	buf.WriteByte('\n')
	file, diags := hclwrite.ParseConfig(buf.Bytes(), "providers.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return at, diags
	}
	block := file.Body().Blocks()[0]
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	if err := reorderBlock(block, names); err != nil {
		return at, err
	}
	inner := block.Body().BuildTokens(nil)
	at.ExprTokens = append(hclwrite.Tokens{exprTokens[0]}, inner...)
	at.ExprTokens = append(at.ExprTokens, exprTokens[len(exprTokens)-1])
	return at, nil
}
