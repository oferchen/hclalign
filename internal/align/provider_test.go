// internal/align/provider_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestProviderNestedBlockOrder(t *testing.T) {
	src := []byte(`provider "aws" {
  nested "b" {}
  assume_role {}
  nested "a" {}
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `provider "aws" {

  assume_role {}

  nested "a" {}

  nested "b" {}
}`
	require.Equal(t, exp, got)
}
