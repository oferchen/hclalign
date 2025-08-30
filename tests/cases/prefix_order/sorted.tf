resource "aws_s3_bucket" "r" {
  bucket = "b"
  bar    = 2
  foo    = 1
}

module "m" {
  source = "./m"
  providers = {
    aws     = aws.us
    azurerm = azurerm
  }
}
