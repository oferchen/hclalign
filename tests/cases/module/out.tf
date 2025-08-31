module "m" {
  source  = "./m"
  version = "1.0"
}

module "complex" {
  z = 0

  provisioner "local-exec" {
    command = "echo hi"
  }

  source  = "./complex"
  version = "1.0"
  providers = {
    b = aws.b
    a = aws.a
  }
  count      = 1
  for_each   = {}
  depends_on = []
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
