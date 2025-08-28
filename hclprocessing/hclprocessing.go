// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var canonicalOrder = []string{"description", "type", "default", "sensitive", "nullable"}

// ReorderAttributes reorders attributes of "variable" blocks into the provided order.
// If order is nil or empty, a canonical order is used.
func ReorderAttributes(file *hclwrite.File, order []string) {
	if len(order) == 0 {
		order = canonicalOrder
	}
	if _, diags := hclwrite.ParseConfig(file.Bytes(), "", hcl.InitialPos); diags.HasErrors() {
		return
	}

	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() != "variable" {
			continue
		}

		reorderVariableBlock(block, order)
	}
}

func reorderVariableBlock(block *hclwrite.Block, order []string) {
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
	if len(tail) > 0 && tail[0].Type == hclsyntax.TokenNewline {
		tail = tail[1:]
	}
	body.Clear()
	body.AppendNewline()

	canonSet := map[string]struct{}{}
	for _, name := range order {
		canonSet[name] = struct{}{}
		if tok, ok := tokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.lead)
			body.SetAttributeRaw(name, tok.expr)
		}
	}

	for _, name := range orderedNames {
		if _, isCanonical := canonSet[name]; isCanonical {
			continue
		}
		if tok, ok := tokensMap[name]; ok {
			body.AppendUnstructuredTokens(tok.lead)
			body.SetAttributeRaw(name, tok.expr)
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
