package hclprocessing_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/hclprocessing"
	"github.com/stretchr/testify/require"
)

func TestReorderAttributes_VariableBlock(t *testing.T) {
	src := `variable "example" {
  default = "v"
  description = "d"
  custom = true
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	hclprocessing.ReorderAttributes(f, nil, false)

	expected := `variable "example" {
  description = "d"
  default     = "v"
  custom      = true
}`
	require.Equal(t, expected, string(f.Bytes()))
}

func TestReorderAttributes_CustomOrder(t *testing.T) {
	src := `variable "example" {
  description = "d"
  default     = "v"
  custom      = true
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	hclprocessing.ReorderAttributes(f, []string{"default", "description"}, false)

	expected := `variable "example" {
  default     = "v"
  description = "d"
  custom      = true
}`
	require.Equal(t, expected, string(f.Bytes()))
}

func TestReorderAttributes_NestedBlocks(t *testing.T) {
	src := `variable "example" {
  default = "v"
  validation {
    condition     = true
    error_message = "msg"
  }
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	hclprocessing.ReorderAttributes(f, nil, false)

	expected := `variable "example" {
  default = "v"
  validation {
    condition     = true
    error_message = "msg"
  }
}`
	require.Equal(t, expected, string(f.Bytes()))
}

func TestReorderAttributes_IgnoresNonVariable(t *testing.T) {
	src := `resource "r" "t" {
  b = 1
  a = 2
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	hclprocessing.ReorderAttributes(f, nil, false)

	require.Equal(t, src, string(f.Bytes()))
}

func TestReorderAttributes_StrictUnknownAttributesAfterKnown(t *testing.T) {
	src := `variable "example" {
  custom      = true
  description = "d"
  type        = string
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	// The provided order intentionally places the unknown attribute first to
	// verify it is moved after known attributes when strict is enabled.
	hclprocessing.ReorderAttributes(f, []string{"custom", "description", "type"}, true)

	expected := `variable "example" {
  description = "d"
  type        = string
  custom      = true
}`
	require.Equal(t, expected, string(f.Bytes()))
}

func TestReorderAttributes_LoosePreservesUnknownPositions(t *testing.T) {
	src := `variable "example" {
  custom      = true
  description = "d"
  type        = string
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	// In loose mode the unknown attribute remains at the top.
	hclprocessing.ReorderAttributes(f, []string{"description", "type"}, false)

	expected := `variable "example" {
  custom      = true
  description = "d"
  type        = string
}`
	require.Equal(t, expected, string(f.Bytes()))
}
