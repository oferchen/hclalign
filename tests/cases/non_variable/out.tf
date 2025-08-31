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
  a = 2
  b = 1
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

dynamic "d" {
  for_each = [1]
  content {
    b = 1
    a = 2
  }
}

lifecycle {
  prevent_destroy       = false
  create_before_destroy = true

  precondition {
    error_message = "pre"
    condition     = true
  }

  postcondition {
    error_message = "post"
    condition     = true
  }
}
