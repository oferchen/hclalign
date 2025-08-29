resource "aws_s3_bucket" "b" {
  depends_on = []
  count      = 1
  for_each   = {}
  provider   = "aws.us"
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
