// internal/align/connection.go
package align

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type connectionStrategy struct{}

func (connectionStrategy) Name() string { return "connection" }

func (connectionStrategy) Align(block *hclwrite.Block, opts *Options) error {
	attrs := block.Body().Attributes()
	names := make([]string, 0, len(attrs))
	allowed := map[string]struct{}{
		"type":                 {},
		"host":                 {},
		"user":                 {},
		"password":             {},
		"private_key":          {},
		"certificate":          {},
		"timeout":              {},
		"script_path":          {},
		"port":                 {},
		"proxy_command":        {},
		"agent":                {},
		"agent_socket":         {},
		"bastion_host":         {},
		"bastion_port":         {},
		"bastion_user":         {},
		"bastion_password":     {},
		"bastion_private_key":  {},
		"bastion_certificate":  {},
		"bastion_agent":        {},
		"bastion_agent_socket": {},
		"host_key":             {},
		"bastion_host_key":     {},
	}
	var unknown []string
	for name := range attrs {
		names = append(names, name)
		if _, ok := allowed[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if opts != nil && opts.Strict && len(unknown) > 0 {
		sort.Strings(unknown)
		return fmt.Errorf("connection: unknown attributes: %s", strings.Join(unknown, ", "))
	}
	sort.Strings(names)
	return reorderBlock(block, names)
}

func init() { Register(connectionStrategy{}) }
