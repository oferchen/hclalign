variable "example" {
  description = "example"
  type        = list(number)
  default = [
    1,
    2,
  ]
  sensitive = true
  nullable  = true
  validation {
    condition     = var.example != ""
    error_message = "first"
  }
  validation {
    condition     = length(var.example) > 1
    error_message = "second"
  }
  foo = "foo"
  bar = "bar"
}
