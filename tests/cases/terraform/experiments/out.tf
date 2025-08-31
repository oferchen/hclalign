terraform {
  required_version = ">= 1.2.0"

  required_providers {}

  backend "s3" {}

  cloud {}
  experiments = ["a"]
  other       = 1
}
