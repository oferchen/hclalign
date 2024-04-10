// example3.tf
variable "example_variable1" {
  validation {
    condition = (
      length(var.example_variable1) < 10
    )
    error_message = "example variable must be less than 10 characters"
  }
  validation {
    condition = (
      length(replace(var.example_variable1, "/[a-z]*/", "")) == 0
    )
    error_message = "example variable must contain only lowercase letters"
  }
  type        = string
  description = "Example variable 1"
}

variable "example_variable2" {
  type        = string
  description = "Example variable 2"
}

variable "example_variable3" {
  default     = "example_variable3_default"
  type        = string
  description = "Example variable 3"
}

variable "example_variable4" {
  type        = map(string)
  default     = {}
  description = "Example variable 4"
}

variable "example_variable5" {
  default = [
    "example_variable5_default1",
    "example_variable5_default2",
    "example_variable5_default3",
    "example_variable5_default4",
    "example_variable5_default5",
  ]
  type        = list(string)
  description = "Example variable 5"
}
