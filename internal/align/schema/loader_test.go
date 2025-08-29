package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const sample = `{
  "provider_schemas": {
    "registry.terraform.io/hashicorp/test": {
      "resource_schemas": {
        "test_thing": {
          "block": {
            "attributes": {
              "foo": {"required": true},
              "bar": {"optional": true},
              "baz": {"computed": true}
            }
          }
        }
      },
      "data_source_schemas": {
        "test_data": {
          "block": {
            "attributes": {
              "id": {"computed": true}
            }
          }
        }
      }
    }
  }
}`

func TestLoad(t *testing.T) {
	r := strings.NewReader(sample)
	schemas, err := Load(r)
	require.NoError(t, err)

	s, ok := schemas["test_thing"]
	require.True(t, ok)
	_, req := s.Required["foo"]
	require.True(t, req)
	_, opt := s.Optional["bar"]
	require.True(t, opt)
	_, comp := s.Computed["baz"]
	require.True(t, comp)
	_, meta := s.Meta["provider"]
	require.True(t, meta)

	ds, ok := schemas["test_data"]
	require.True(t, ok)
	_, comp2 := ds.Computed["id"]
	require.True(t, comp2)
}
