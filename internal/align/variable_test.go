// /internal/align/variable_test.go
package align

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestVariableCustomOrder(t *testing.T) {
	src := `variable "example" {
  description = "d"
  default     = "v"
  custom      = true
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	if err := Apply(f, &Options{Order: []string{"default", "description"}}); err != nil {
		t.Fatalf("align: %v", err)
	}
	expected := `variable "example" {
  default     = "v"
  description = "d"
  custom      = true
}`
	if got := string(f.Bytes()); got != expected {
		t.Fatalf("output mismatch:\n-- got --\n%s\n-- want --\n%s", got, expected)
	}
}

func TestVariableCustomOrderUnknownAfterCanonical(t *testing.T) {
	src := `variable "example" {
  custom      = true
  type        = string
  description = "d"
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	if err := Apply(f, &Options{Order: []string{"description", "type"}}); err != nil {
		t.Fatalf("align: %v", err)
	}
	expected := `variable "example" {
  description = "d"
  type        = string
  custom      = true
}`
	if got := string(f.Bytes()); got != expected {
		t.Fatalf("output mismatch:\n-- got --\n%s\n-- want --\n%s", got, expected)
	}
}
