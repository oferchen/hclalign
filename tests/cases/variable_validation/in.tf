variable "valid" {
  description = "desc"
  type        = string
  default     = "ok"
  sensitive   = true
  nullable    = false
  validation {
    condition     = true
    error_message = "msg1"
  }
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
}
