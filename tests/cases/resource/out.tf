resource "aws_s3_bucket" "b" {
  provider   = "aws.us"
  count      = 1
  for_each   = {}
  depends_on = []
  list_attr = [
    "x",
    "y",
  ]
  acl = "private"
  tags = [
    "a",
    "b",
  ]
  bucket = "b"
  id     = "id"

  provisioner "local-exec" {
    command = "echo"
  }

  lifecycle {
    prevent_destroy = true
  }
}
