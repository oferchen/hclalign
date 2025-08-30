variable "valid" {
  validation {
    condition     = true
    error_message = "msg1"
  }
  nullable    = false
  sensitive   = true
  default     = "ok"
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
  description = "desc"
  type        = string
}
