variable "multi" {
  description = "desc"
  type        = string
  default     = "ok"
  bar         = 2
  foo         = 1
  baz         = 3
  validation {
    condition     = true
    error_message = "msg1"
  }
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
}
