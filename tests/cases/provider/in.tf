provider "aws" {
  # region comment
  region = "us-east-1"
  # access key comment
  access_key = "foo"
  # secret key comment
  secret_key = "bar"
  # alias comment
  alias = "east"

  # assume role block
  assume_role {
    role_arn = "arn:aws:iam::123456789012:role/test"
  }
}
