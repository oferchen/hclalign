// internal/align/output_test.go
package align_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	alignpkg "github.com/oferchen/hclalign/internal/align"
	"github.com/stretchr/testify/require"
)

func TestOutputAttributeOrder(t *testing.T) {
	src := []byte(`output "example" {
  depends_on  = [var.x]
  value       = var.v
  ephemeral   = true
  description = "desc"
  sensitive   = true
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `output "example" {
  description = "desc"
  value       = var.v
  sensitive   = true
  ephemeral   = true
  depends_on  = [var.x]
}`
	require.Equal(t, exp, got)
}

func TestOutputBlockOrder(t *testing.T) {
	src := []byte(`output "example" {
  postcondition {
    condition     = true
    error_message = "post1"
  }
  precondition {
    condition     = true
    error_message = "pre1"
  }
  other {}
  postcondition {
    condition     = false
    error_message = "post2"
  }
  precondition {
    condition     = false
    error_message = "pre2"
  }
}`)
	file, diags := hclwrite.ParseConfig(src, "in.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	require.NoError(t, alignpkg.Apply(file, &alignpkg.Options{}))
	got := string(file.Bytes())
	exp := `output "example" {

  precondition {
    condition     = true
    error_message = "pre1"
  }

  precondition {
    condition     = false
    error_message = "pre2"
  }

  postcondition {
    condition     = true
    error_message = "post1"
  }

  postcondition {
    condition     = false
    error_message = "post2"
  }

  other {}
}`
	require.Equal(t, exp, got)
}
