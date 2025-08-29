variable "example" {
  foo         = "foo"
  description = "example"
  bar         = "bar"
  type        = number
  default     = 1
  sensitive   = true
  nullable    = false
}
