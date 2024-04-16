variable "example_var_inline_comment" { default = "corner { } case" // example8.tf
  # Problematic corner case inline comment
  // Corner case? # what about this? /* about now? */
  /* multi line
comment */ description = "An example variable with an inline comment" type = string # Just another inline 
  // different tyoe of comment
  validation = { key = value } // this is screwed up
  # how about now?
  # and this one?
}

