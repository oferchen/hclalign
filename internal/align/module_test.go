// internal/align/module_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestModuleProvisionerAndProviders(t *testing.T) {
	src := []byte(`module "example" {
  z = 1

  provisioner "local-exec" {}

  source    = "./m"
  providers = {
    b = aws.b
    a = aws.a
  }
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `module "example" {
  z = 1

  provisioner "local-exec" {}

  source = "./m"
  providers = {
    b = aws.b
    a = aws.a
  }
}`
	require.Equal(t, exp, got)
}

func TestModuleProvidersPrefixOrder(t *testing.T) {
	src := []byte(`module "example" {
  source    = "./m"
  providers = {
    b = aws.b
    a = aws.a
  }
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	got := string(file.Bytes())
	exp := `module "example" {
  source = "./m"
  providers = {
    a = aws.a
    b = aws.b
  }
}`
	require.Equal(t, exp, got)
}
