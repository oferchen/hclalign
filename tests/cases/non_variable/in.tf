resource "r" "t" {
  b = 1
  a = 2
}

data "d" "t" {
  b = 1
  a = 2
}

provider "p" {
  b = 1
  a = 2
}

module "m" {
  b = 1
  a = 2
}

output "o" {
  b = 1
  a = 2
}

locals {
  b = 1
  a = 2
}

terraform {
  b = 1
  a = 2
}

moved {
  from = aws_instance.foo
  to   = aws_instance.bar
}

import {
  id = "i-123"
  to = aws_instance.example
}
