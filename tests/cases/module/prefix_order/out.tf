module "example" {
  source = "./m"
  providers = {
    a = aws.a
    b = aws.b
  }
}
