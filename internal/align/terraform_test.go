// /internal/align/terraform_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestTerraformAttributeOrderAndBlocks(t *testing.T) {
	src := []byte(`terraform {
  backend "s3" {}
  required_providers {}
  experiments = ["test"]
  required_version = ">= 1.2.0"
  cloud {}
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `terraform {
  required_version = ">= 1.2.0"
  experiments      = ["test"]

  backend "s3" {}

  required_providers {}

  cloud {}
}`
	require.Equal(t, exp, got)
}

func TestTerraformRequiredProvidersSorting(t *testing.T) {
	src := []byte(`terraform {
    required_providers {
      b = {}
      a = {}
    }
  }`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	exp := `terraform {

  required_providers {
    b = {}
    a = {}
  }
}`
	require.Equal(t, exp, string(file.Bytes()))
}
