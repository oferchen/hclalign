provider "aws" {
  # alias comment
  alias = "east"
  # access key comment
  access_key = "foo"
  # region comment
  region = "us-east-1"
  # secret key comment
  secret_key = "bar"

  # assume role block
  assume_role {
    role_arn = "arn:aws:iam::123456789012:role/test"
  }

  # nested a comment
  nested "a" {
    v = 1
  }

  # nested b comment
  nested "b" {
    v = 2
  }
}
