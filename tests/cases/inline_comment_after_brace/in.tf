// example8.tf
variable "example_var_inline_comment" { # Problematic corner case inline comment
default = "corner { } case"
/* multi line
comment */
  description = "An example variable with an inline comment" # Just another inline 
// different tyoe of comment
type = string
}
