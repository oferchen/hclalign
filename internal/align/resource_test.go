// internal/align/resource_test.go
package align_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	alignschema "github.com/oferchen/hclalign/internal/align/schema"
	"github.com/stretchr/testify/require"
)

func TestSchemaAwareOrder(t *testing.T) {
	src := []byte(`resource "test_thing" "ex" {
  provider   = "p"
  baz        = 3
  bar        = 2
  depends_on = []
  foo        = 1
  random     = 4
}`)

	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	sch := &alignpkg.Schema{
		Required: map[string]struct{}{"foo": {}},
		Optional: map[string]struct{}{"bar": {}},
		Computed: map[string]struct{}{"baz": {}},
		Meta:     map[string]struct{}{"provider": {}, "depends_on": {}, "count": {}, "for_each": {}},
	}
	schemas := map[string]*alignpkg.Schema{"test_thing": sch}

	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{Schemas: schemas}))

	got := string(file.Bytes())
	exp := `resource "test_thing" "ex" {
  provider   = "p"
  depends_on = []
  foo        = 1
  bar        = 2
  baz        = 3
  random     = 4
}`
	require.Equal(t, exp, got)
}

func TestProviderSchemaResourceOrdering(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "providers-schema.json")
	schemas, err := alignschema.LoadFile(path)
	require.NoError(t, err)

	src := []byte(`resource "aws_s3_bucket" "b" {
  tags   = {}
  id     = "id"
  bucket = "b"
  acl    = "private"
}

resource "null_resource" "n" {
  id       = "nid"
  triggers = {}
}`)

	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{Schemas: schemas}))

	got := string(file.Bytes())
	exp := `resource "aws_s3_bucket" "b" {
  bucket = "b"
  acl    = "private"
  tags   = {}
  id     = "id"
}

resource "null_resource" "n" {
  triggers = {}
  id       = "nid"
}`
	require.Equal(t, exp, got)
}
