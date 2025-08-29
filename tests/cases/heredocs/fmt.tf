variable "heredoc" {
  default     = <<EOF
line1
line2
EOF
  description = <<-EOT
  indented
  text
EOT
  type        = string
}