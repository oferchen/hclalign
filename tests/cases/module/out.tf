module "m" {
  source  = "./m"
  version = "1.0"
}

module "complex" {
  source     = "./complex"
  version    = "1.0"
  count      = 1
  for_each   = {}
  providers  = {}
  depends_on = []
  a          = 1
  b          = 2
  c          = 3
  z          = 0

  foo {
    x = 1
  }
}

module "vars_only" {
  a = 1
  b = 2
  c = 3
}
