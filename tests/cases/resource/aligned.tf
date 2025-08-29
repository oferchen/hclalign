resource "aws_s3_bucket" "b" {
  provider   = "aws.us"
  count      = 1
  for_each   = {}
  depends_on = []
  bucket     = "b"
  acl        = "private"
  tags       = {}
  id         = "id"
  bar        = "bar"
  foo        = "foo"
}

resource "null_resource" "n" {
  triggers = {}
  id       = "nid"
}
