variable "example" {
  default = [
    1,
    2,
  ]
  validation {
    condition     = var.example != ""
    error_message = "first"
  }
  nullable = true
  description = "example"
  validation {
    condition     = length(var.example) > 1
    error_message = "second"
  }
  sensitive = true
  type = list(number)
  foo="foo"
  bar = "bar"
}
