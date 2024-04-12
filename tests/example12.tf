// Top-level single-line comment

/*
Top-level multi-line comment
covering several lines
*/

variable "edge_case_demo" { # Inline comment next to the block opener
  default = [ "item1", "item2" ] /* Inline comment next to an array */

  // Comment before nested block
  nested_block {
    // Nested block single-line comment
    description = "A description here" /* End-of-line comment */

    /* Multi-line comment before nested detail
       spanning multiple lines */
    detail {
      more_detail = false // End-of-line comment in nested detail
      /* Start of multi-line comment in array
         continuation of comment */
      items = [1, 2, 3] // Inline comment after array
    }
  }

  type = "complex" // Comment after simple attribute

  /* Multi-line comment before heredoc attribute */
  message = <<-EOT
Multi-line heredoc content here,
with special characters: {}, [], #, etc.
EOT
// This heredoc ends with an inline comment

  // Comment before sensitive attribute
  sensitive = true /* Inline comment after sensitive attribute */

  validation = regex("^[a-zA-Z]+${var.count}") # Inline comment with interpolation

  /* Multi-line comment block before final attribute */
  configuration = {
    parameters = "Yes" /* Inline comment inside map */
    enabled = true // Single-line comment inside map
  }

  // This is a comment at the end of the block
}
