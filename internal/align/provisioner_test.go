// internal/align/provisioner_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestProvisionerAttributeOrderAndComments(t *testing.T) {
	src := []byte(`provisioner "local-exec" {
  when = "destroy" // when inline
  foo = "bar" // foo inline
  on_failure = "continue" // on_failure inline
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	got := string(file.Bytes())
	exp := `provisioner "local-exec" {
  on_failure = "continue" // on_failure inline
  when       = "destroy" // when inline
  foo        = "bar" // foo inline
}`
	require.Equal(t, exp, got)
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	require.Equal(t, exp, string(file.Bytes()))
}
