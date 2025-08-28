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
	tokensMap := make(map[string]hclwrite.Tokens)
	for name, attr := range attrs {
		tokensMap[name] = attr.Expr().BuildTokens(nil)
	}

	// Capture original attribute ordering.
	orderedNames := attributeOrder(body, attrs)

	for name := range attrs {
		body.RemoveAttribute(name)
	}

	canonSet := map[string]struct{}{}
	for _, name := range order {
		canonSet[name] = struct{}{}
		if tokens, ok := tokensMap[name]; ok {
			body.SetAttributeRaw(name, tokens)
		}
	}

	for _, name := range orderedNames {
		if _, isCanonical := canonSet[name]; isCanonical {
			continue
		}
		if tokens, ok := tokensMap[name]; ok {
			body.SetAttributeRaw(name, tokens)
		}
	}

	for _, nb := range nested {
		body.AppendBlock(nb)
	}
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
