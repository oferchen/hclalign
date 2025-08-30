variable "complex" {
  description = "desc"
  type        = list(string)
  default     = ["a", "b"]
  sensitive   = true
  nullable    = false
  validation {
    condition     = true
    error_message = "msg"
  }
  custom = true
}