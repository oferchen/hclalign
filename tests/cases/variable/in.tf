variable "example" {
  default=1
  bar = "bar"
  description = "example"
  foo="foo"
  type = number

  validation {
    condition     = length(var.example) > 0
    error_message = "not empty"
  }
}
