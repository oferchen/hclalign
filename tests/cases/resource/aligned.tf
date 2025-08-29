resource "aws_s3_bucket" "b" {
  bucket = "b"
  acl    = "private"
  tags   = {}
}
