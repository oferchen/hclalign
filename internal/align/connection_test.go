// internal/align/connection_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestConnectionAttributeOrder(t *testing.T) {
	src := []byte(`resource "r" "t" {
  connection {
    user    = "u"
    host    = "h"
    type    = "ssh"
    timeout = "1s"
  }
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `resource "r" "t" {

  connection {
    host    = "h"
    timeout = "1s"
    type    = "ssh"
    user    = "u"
  }
}`
	require.Equal(t, exp, got)
}
