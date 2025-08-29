// internal/hclalign/hclalign_test.go
package hclalign_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/internal/hclalign"
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

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

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

	require.NoError(t, hclalign.ReorderAttributes(f, []string{"default", "description"}, false))

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

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

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

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

	require.Equal(t, src, string(f.Bytes()))
}

func TestReorderAttributes_StrictUnknownAttrError(t *testing.T) {
	src := `variable "example" {
  custom      = true
  description = "d"
  type        = string
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	err := hclalign.ReorderAttributes(f, []string{"custom", "description", "type"}, true)
	require.Error(t, err)
}

func TestReorderAttributes_StrictUnknownAttrWithCanonical(t *testing.T) {
	src := `variable "example" {
  description = "d"
  type        = string
  default     = "v"
  sensitive   = true
  nullable    = false
  custom      = true
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	err := hclalign.ReorderAttributes(f, nil, true)
	require.Error(t, err)
}

func TestReorderAttributes_LooseRetainsUnknownOrder(t *testing.T) {
	src := `variable "example" {
  custom      = true
  type        = string
  description = "d"
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	require.NoError(t, hclalign.ReorderAttributes(f, []string{"description", "type"}, false))

	expected := `variable "example" {
  custom      = true
  description = "d"
  type        = string
}`
	require.Equal(t, expected, string(f.Bytes()))
}

func TestReorderAttributes_FirstAttrSameLine(t *testing.T) {
	src := `variable "example" { default = "v" }`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

	require.Equal(t, src, string(f.Bytes()))
}

func TestReorderAttributes_OnlyNestedBlocks(t *testing.T) {
	src := `variable "example" {
  validation {
    condition = true
  }
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

	require.Equal(t, src, string(f.Bytes()))
}

func TestReorderAttributes_DefaultBlockNestedType(t *testing.T) {
	src := `variable "example" {
  type    = string
  default = {
    type = string
  }
}`
	f, diags := hclwrite.ParseConfig([]byte(src), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	require.NoError(t, hclalign.ReorderAttributes(f, nil, false))

	expected := `variable "example" {
  type = string
  default = {
    type = string
  }
}`
	require.Equal(t, expected, string(f.Bytes()))
}
