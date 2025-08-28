variable "valid" {
  type    = string
  default = "ok"
  validation {
    condition     = true
    error_message = "msg1"
  }
  validation {
    condition     = 2 > 1
    error_message = "msg2"
  }
}