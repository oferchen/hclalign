variable "interpolation" {
  type    = string
  default = "${var.foo}"
}

variable "directive" {
  type    = string
  default = <<-EOT
%{if var.bar}
${var.bar}
%{endif}
EOT
}