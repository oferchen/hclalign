// /internal/align/terraform.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ihcl "github.com/oferchen/hclalign/internal/hcl"
)

type terraformStrategy struct{}

func (terraformStrategy) Name() string { return "terraform" }

func (terraformStrategy) Align(block *hclwrite.Block, opts *Options) error {
	body := block.Body()

	attrs := body.Attributes()
	blocks := body.Blocks()

	canonical := []string{"required_version", "experiments", "required_providers", "backend", "cloud"}
	canonSet := make(map[string]struct{}, len(canonical))
	for _, n := range canonical {
		canonSet[n] = struct{}{}
	}

	if opts != nil && opts.Strict {
		required := []string{"required_version", "required_providers", "backend", "cloud"}
		var missing []string
		for _, n := range required {
			if _, ok := attrs[n]; ok {
				continue
			}
			found := false
			for _, b := range blocks {
				if b.Type() == n {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			return fmt.Errorf("terraform: missing attributes or blocks: %s", strings.Join(missing, ", "))
		}
	}

	for _, nb := range blocks {
		if nb.Type() == "required_providers" && opts != nil && opts.Strict {
			attrs := nb.Body().Attributes()
			names := make([]string, 0, len(attrs))
			for name := range attrs {
				names = append(names, name)
			}
			sort.Strings(names)
			if err := reorderBlock(nb, names); err != nil {
				return err
			}
		}
	}

	tokens := body.BuildTokens(nil)
	newline := ihcl.DetectLineEnding(tokens)
	trailingComma := ihcl.HasTrailingComma(tokens)

	order := ihcl.AttributeOrder(body, attrs)

	attrTokens := map[string]ihcl.AttrTokens{}
	for name, attr := range attrs {
		attrTokens[name] = ihcl.ExtractAttrTokens(attr)
		body.RemoveAttribute(name)
	}

	for _, b := range blocks {
		body.RemoveBlock(b)
	}

	type item struct {
		name   string
		block  *hclwrite.Block
		isAttr bool
	}

	var cloudBlock, backendBlock, requiredProviders *hclwrite.Block
	var otherBlocks []*hclwrite.Block
	for _, b := range blocks {
		switch b.Type() {
		case "required_providers":
			requiredProviders = b
		case "backend":
			backendBlock = b
		case "cloud":
			cloudBlock = b
		default:
			otherBlocks = append(otherBlocks, b)
		}
	}

	var otherAttrs []string
	for _, name := range order {
		if name == "required_version" || name == "experiments" {
			continue
		}
		otherAttrs = append(otherAttrs, name)
	}

	var items []item
	if _, ok := attrTokens["required_version"]; ok {
		items = append(items, item{name: "required_version", isAttr: true})
	}
	if _, ok := attrTokens["experiments"]; ok {
		items = append(items, item{name: "experiments", isAttr: true})
	}
	if cloudBlock != nil {
		items = append(items, item{block: cloudBlock})
	}
	if backendBlock != nil {
		items = append(items, item{block: backendBlock})
	}
	if requiredProviders != nil {
		items = append(items, item{block: requiredProviders})
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
