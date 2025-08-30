variable "multi" {
  bar         = 2
  validation {
    condition     = true
    error_message = "msg1"
  }
  default     = "ok"
  foo         = 1
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
  description = "desc"
  baz         = 3
  type        = string
}
