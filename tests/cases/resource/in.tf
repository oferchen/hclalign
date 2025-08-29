resource "aws_s3_bucket" "b" {
  tags   = {}
  id     = "id"
  bucket = "b"
  acl    = "private"
}

resource "null_resource" "n" {
  id       = "nid"
  triggers = {}
}
