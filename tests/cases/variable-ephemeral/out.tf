variable "example" {
  description = "example"
  type        = number
  default     = 1
  sensitive   = true
  nullable    = true
  ephemeral   = true
}
