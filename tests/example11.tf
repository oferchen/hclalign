// This is a single-line comment at the top of the file

/* This is a multi-line comment block
   that spans multiple lines at the top of the file
*/

variable "complex_example" { # Inline comment next to the block opener
  /* Multi-line comment before an attribute */
  default = [ "list", "of", "values" ] // Inline comment at line end

  // Single-line comment before nested block
  nested_block {
    // Comment inside nested block
    description = "Nested block inside a variable block" # Inline comment inside nested
    detail = {
      more_detail = true // Comment at the end of nested nested block
      /* Multi-line comment inside nested detail block
         that spans multiple lines */
      even_more_detail = [1, 2, 3] /* Inline block comment at line end */
    }
  }

  type = "string" /* Multi-line comment after an attribute
  that continues here. */

  // Another single-line comment before an attribute
  sensitive = false /* Inline comment after attribute with space */

  validation = regex("^example\\d+$") # Another type of inline comment
  # Standalone single-line comment between attributes

  /* This is a multi-line comment block
     between attributes */

  description = <<-EOF
  Here is a heredoc string that might contain a lot of text,
  including "quotes" and # hash signs that shouldn't start a comment.
  EOF // End of heredoc string comment

  another_nested_block {
    content = "Just some more nested content" /* End of block comment */
    /* Start of block comment
       that shouldn't affect the next line */
    toggle = true
  }

  // Final comment at the end of the block
}
