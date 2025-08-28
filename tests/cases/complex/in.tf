variable "complex" {
  custom = true
  nullable = false
  default = ["a", "b"]
  validation {
    condition     = true
    error_message = "msg"
  }
  description = "desc"
  sensitive = true
  type = list(string)
}