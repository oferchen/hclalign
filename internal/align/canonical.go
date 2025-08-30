// internal/align/canonical.go
package align

import "github.com/oferchen/hclalign/config"

// CanonicalBlockAttrOrder maps block types to their default attribute ordering.
var CanonicalBlockAttrOrder = map[string][]string{
	"variable": config.CanonicalOrder,
	"output":   {"description", "value", "sensitive", "depends_on"},
	"module":   {"source", "version", "providers", "count", "for_each", "depends_on"},
	"provider": {"alias"},
	"resource": {"provider", "count", "for_each", "depends_on"},
	"data":     {"provider", "count", "for_each", "depends_on"},
}
