module "m" {
  source  = "./m"
  version = "1.0"
}

module "complex" {
  source     = "./complex"
  version    = "1.0"
  providers  = {}
  count      = 1
  for_each   = {}
  depends_on = []
  z          = 0
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
