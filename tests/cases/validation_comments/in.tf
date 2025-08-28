variable "vc" {
  default = ""
  type    = string

  # ensure non-empty
  validation {
    condition     = length(var.vc) > 0
    error_message = "msg"
  }
}