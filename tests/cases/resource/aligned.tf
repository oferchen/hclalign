resource "aws_s3_bucket" "b" {
  bucket = "b"
  acl    = "private"
  tags   = {}
  id     = "id"
}

resource "null_resource" "n" {
  triggers = {}
  id       = "nid"
}
