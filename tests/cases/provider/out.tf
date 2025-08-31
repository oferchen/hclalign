provider "aws" {
  # alias comment
  alias = "east"
  # region comment
  region = "us-east-1"
  # access key comment
  access_key = "foo"
  # secret key comment
  secret_key = "bar"

  # nested b comment
  nested "b" {
    v = 2
  }

  # assume role block
  assume_role {
    role_arn = "arn:aws:iam::123456789012:role/test"
  }

  # nested a comment
  nested "a" {
    v = 1
  }
}
