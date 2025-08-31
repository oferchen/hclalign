// internal/align/data_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestDataPrefixOrder(t *testing.T) {
	src := []byte(`data "x" "y" {
  z = 1
  a = 2
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	exp := `data "x" "y" {
  a = 2
  z = 1
}`
	require.Equal(t, exp, string(file.Bytes()))
}
