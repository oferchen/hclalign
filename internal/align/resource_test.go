package align

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

	sch := &Schema{
		Required: map[string]struct{}{"foo": {}},
		Optional: map[string]struct{}{"bar": {}},
		Computed: map[string]struct{}{"baz": {}},
		Meta:     map[string]struct{}{"provider": {}, "depends_on": {}, "count": {}, "for_each": {}},
	}
	schemas := map[string]*Schema{"test_thing": sch}

	require.NoError(t, Apply(file, &Options{Schemas: schemas}))

	got := string(file.Bytes())
	exp := `resource "test_thing" "ex" {
  foo        = 1
  bar        = 2
  baz        = 3
  depends_on = []
  provider   = "p"
  random     = 4
}`
	require.Equal(t, exp, got)
}
