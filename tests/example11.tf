variable "complex_example" { # Inline comment next to the block opener
  default = [ "list", "of", "values" ] // Inline comment at line end

  nested_block {
    description = "Nested block inside a variable block" # Inline comment inside nested
    detail = {
      more_detail = true // End of nested nested block comment
      even_more_detail = [1, 2, 3] /* Inline block comment at line end */
    }
  }

  type = "string" /* Multi-line comment after an attribute */

  sensitive = false /* Inline comment after attribute */

  validation = regex("^example\\d+$") # Inline comment after attribute

  description = <<-EOF
  Here is a heredoc string that might contain a lot of text,
  including "quotes" and # hash signs that shouldn't start a comment.
  EOF

  another_nested_block {
    content = "Just some more nested content" /* End of block comment */
    toggle = true
  }
}
