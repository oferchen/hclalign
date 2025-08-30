resource "aws_s3_bucket" "b" {
  list_attr = [
    "x",
    "y",
  ]
  provisioner "local-exec" {
    command = "echo"
  }
  acl = "private"
  tags = [
    "a",
    "b",
  ]
  provider = "aws.us"
  for_each = {}
  lifecycle {
    prevent_destroy = true
  }
  count      = 1
  depends_on = []
  bucket     = "b"
  id         = "id"
}
