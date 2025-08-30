// internal/align/dynamic_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestDynamicAttributeOrderAndComments(t *testing.T) {
	src := []byte(`dynamic "x" {
  foo = 1 // foo inline
  iterator = "it" // iterator inline
  for_each = var.list // for_each inline
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	got := string(file.Bytes())
	exp := `dynamic "x" {
  for_each = var.list // for_each inline
  iterator = "it" // iterator inline
  foo      = 1 // foo inline
}`
	require.Equal(t, exp, got)
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	require.Equal(t, exp, string(file.Bytes()))
}
