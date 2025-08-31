variable "example" {
  ephemeral  = true
  nullable   = true
  description = "example"
  type        = number
  default     = 1
  sensitive   = true
}
