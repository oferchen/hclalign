module "example" {
  providers = {
    b = aws.b
    a = aws.a
  }
  source = "./m"
}
