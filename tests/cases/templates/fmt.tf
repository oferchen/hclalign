variable "interpolation" {
  default = "${var.foo}"
  type    = string
}

variable "directive" {
  default = <<-EOT
%{if var.bar}
${var.bar}
%{endif}
EOT
  type    = string
}