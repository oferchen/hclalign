module "m" {
  version = "1.0"
  source  = "./m"
}

module "complex" {
  z = 0

  provisioner "local-exec" {
    command = "echo hi"
  }

  for_each  = {}
  source    = "./complex"
  providers = {
    b = aws.b
    a = aws.a
  }
  depends_on = []
  version    = "1.0"
  count      = 1
  a          = 1
  c          = 3
  b          = 2

  foo {
    x = 1
  }
}

module "vars_only" {
  c = 3
  b = 2
  a = 1
}
