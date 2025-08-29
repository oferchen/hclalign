resource "r" "t" {
  b = 1
  a = 2
}

data "d" "t" {
  b = 1
  a = 2
}

provider "p" {
  a = 2
  b = 1
}

module "m" {
  a = 2
  b = 1
}

output "o" {
  b = 1
  a = 2
}

locals {
  a = 2
  b = 1
}

terraform {
  a = 2
  b = 1
}

moved {
  from = aws_instance.foo
  to   = aws_instance.bar
}

import {
  id = "i-123"
  to = aws_instance.example
}

dynamic "d" {
  for_each = [1]

  content {
    b = 1
    a = 2
  }
}

lifecycle {
  create_before_destroy = true
  prevent_destroy       = false

  precondition {
    error_message = "pre"
    condition     = true
  }

  postcondition {
    error_message = "post"
    condition     = true
  }
}
