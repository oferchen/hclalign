// internal/align/terraform_test.go
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
  experiments = ["a"]
  required_providers {}
  required_version = ">= 1.2.0"
  other = 1
  cloud {}
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `terraform {
  required_version = ">= 1.2.0"

  required_providers {}

  backend "s3" {}

  cloud {}
  experiments = ["a"]
  other       = 1
}`
	require.Equal(t, exp, got)
}

func TestTerraformRequiredProvidersSorting(t *testing.T) {
	src := []byte(`terraform {
  required_providers {
    # provider b
    b = {}
    # provider a
    a = {}
  }
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	exp := `terraform {

  required_providers {
    # provider b
    b = {}
    # provider a
    a = {}
  }
}`
	require.Equal(t, exp, string(file.Bytes()))
}

func TestTerraformBlocksOrderWithoutExperiments(t *testing.T) {
	src := []byte(`terraform {
  backend "s3" {}
  required_providers {}
  required_version = ">= 1.2.0"
  other = 1
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	exp := `terraform {
  required_version = ">= 1.2.0"

  required_providers {}

  backend "s3" {}
  other = 1
}`
	require.Equal(t, exp, string(file.Bytes()))
}
