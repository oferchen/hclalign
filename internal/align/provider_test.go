// internal/align/provider_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestProviderNestedBlocksPreserveOrder(t *testing.T) {
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

  nested "b" {}

  assume_role {}

  nested "a" {}
}`
	require.Equal(t, exp, got)
}

func TestProviderAttributeOrder(t *testing.T) {
	src := []byte(`provider "aws" {
  region  = "us-east-1"
  alias   = "west"
  profile = "default"
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `provider "aws" {
  alias   = "west"
  profile = "default"
  region  = "us-east-1"
}`
	require.Equal(t, exp, got)
}

func TestProviderAttributeOrderNoAlias(t *testing.T) {
	src := []byte(`provider "aws" {
  region  = "us-east-1"
  profile = "default"
  version = "~> 4.0"
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `provider "aws" {
  profile = "default"
  region  = "us-east-1"
  version = "~> 4.0"
}`
	require.Equal(t, exp, got)
}

func TestProviderAttributePrefixOrder(t *testing.T) {
	src := []byte(`provider "aws" {
  zz = 1
  aa = 2
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{PrefixOrder: true}))
	got := string(file.Bytes())
	exp := `provider "aws" {
  aa = 2
  zz = 1
}`
	require.Equal(t, exp, got)
}
