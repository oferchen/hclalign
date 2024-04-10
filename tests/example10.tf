// example10.tf

// Single-line comment
variable "example_variable_1" {
  description = "Example variable 1"
}

// Another single-line comment
variable "example_variable_2" {
  description = "Example variable 2"
}

// Multi-line comment
/*
  This is a multi-line comment.
  It spans multiple lines.
*/
variable "example_variable_3" {
  description = "An example variable"
  default = {
    key1 = "value1"
    key2 = "value2"
  }
}

// Variable with validation
variable "example_variable_4" {
  type = object
  object1 = {
    key1 = "value1"
    key2 = "value2"
  }
  validation {
    condition = length(var.example_variable_4) < 10
    error_message = "Variable length must be less than 10"
  }
}

// Variable with default list
variable "example_variable_5" {
  default = [
    "example_variable5_default1",
    "example_variable5_default2",
    "example_variable5_default3"
  ]
}
