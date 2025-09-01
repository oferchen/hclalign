// internal/align/strategy.go
package align

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Options struct {
	Order   []string
	Schemas map[string]*Schema

	Schema *Schema

	Types map[string]struct{}
}

type Schema struct {
	Required map[string]struct{}
	Optional map[string]struct{}
	Computed map[string]struct{}
	Meta     map[string]struct{}

	RequiredOrder []string
	OptionalOrder []string
	ComputedOrder []string

	Blocks      map[string]*Schema
	BlocksOrder []string
}

type Strategy interface {
	Name() string

	Align(block *hclwrite.Block, opts *Options) error
}

var registry = map[string]Strategy{}

func Register(s Strategy) {
	registry[s.Name()] = s
}

func Apply(file *hclwrite.File, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	return applyBody(file.Body(), opts)
}

func applyBody(body *hclwrite.Body, opts *Options) error {
	for _, b := range body.Blocks() {
		sub := *opts
		sub.Schema = nil
		if len(b.Labels()) > 0 && opts.Schemas != nil {
			typ := b.Labels()[0]
			if s, ok := opts.Schemas[typ]; ok {
				sub.Schema = s
			} else {
				for k, s := range opts.Schemas {
					if strings.HasSuffix(k, "/"+typ) {
						sub.Schema = s
						break
					}
				}
			}
		} else if opts.Schema != nil {
			if s, ok := opts.Schema.Blocks[b.Type()]; ok {
				sub.Schema = s
			}
		}

		if strategy, ok := registry[b.Type()]; ok {
			if opts.Types != nil {
				if _, ok := opts.Types[b.Type()]; !ok {
					if err := applyBody(b.Body(), &sub); err != nil {
						return err
					}
					continue
				}
			}
			if err := strategy.Align(b, &sub); err != nil {
				return err
			}
		}
		if err := applyBody(b.Body(), &sub); err != nil {
			return err
		}
	}
	return nil
}
