variable "example" {
  description = "example"
  type        = number
  default     = 1
  bar         = "bar"
  foo         = "foo"

  validation {
    condition     = length(var.example) > 0
    error_message = "not empty"
  }
}
