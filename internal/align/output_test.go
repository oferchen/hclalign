// internal/align/output_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestOutputAttributeOrder(t *testing.T) {
	src := []byte(`output "ephemeral" {
  value      = var.v
  depends_on = [var.x]
  ephemeral  = true
  sensitive  = false
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `output "ephemeral" {
  value      = var.v
  sensitive  = false
  depends_on = [var.x]
  ephemeral  = true
}`
	require.Equal(t, exp, got)
}
