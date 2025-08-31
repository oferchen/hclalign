// internal/align/canonical_test.go
package align_test

import (
	"testing"

	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestCanonicalTerraformOrder(t *testing.T) {
	exp := []string{"required_version", "required_providers", "backend", "cloud"}
	require.Equal(t, exp, alignpkg.CanonicalBlockAttrOrder["terraform"])
}
