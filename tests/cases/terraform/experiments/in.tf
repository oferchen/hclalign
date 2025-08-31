terraform {
  backend "s3" {}
  experiments = ["a"]
  required_providers {}
  required_version = ">= 1.2.0"
  other = 1
  cloud {}
}
