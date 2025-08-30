variable "heredoc" {
  description = <<-EOT
  indented
  text
EOT
  type        = string
  default     = <<EOF
line1
line2
EOF
}