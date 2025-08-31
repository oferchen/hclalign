resource "aws_s3_bucket" "r" {
  bucket = "b"
  foo    = 1
  bar    = 2
}

module "m" {
  source = "./m"
  providers = {
    azurerm = azurerm
    aws     = aws.us
  }
  a = 1
  b = 2
  c = 3
}
