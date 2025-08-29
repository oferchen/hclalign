resource "aws_s3_bucket" "b" {
  tags   = {}
  bucket = "b"
  acl    = "private"
}
