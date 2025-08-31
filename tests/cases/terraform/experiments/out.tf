terraform {
  required_version = ">= 1.2.0"
  experiments      = ["a"]

  required_providers {}

  backend "s3" {}

  cloud {}
  other = 1
}
