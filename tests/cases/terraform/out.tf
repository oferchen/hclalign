terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      version = "~> 4.0"
      source  = "hashicorp/aws"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  backend "s3" {
    region = "us-east-1"
  }

  cloud {
    organization = "hashicorp"
  }
}
