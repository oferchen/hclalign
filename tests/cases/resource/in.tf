resource "aws_s3_bucket" "b" {
  bar        = "bar"
  provider   = "aws.us"
  for_each   = {}
  foo        = "foo"
  depends_on = []
  bucket     = "b"
  count      = 1
  acl        = "private"
  tags       = {}
  id         = "id"
}

resource "null_resource" "n" {
  id       = "nid"
  triggers = {}
}
