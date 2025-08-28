variable "valid" {
  default = "ok"
  validation {
    condition     = true
    error_message = "msg1"
  }
  type = string
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
}