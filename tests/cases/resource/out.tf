resource "aws_s3_bucket" "b" {
  provider   = "aws.us"
  count      = 1
  for_each   = {}
  depends_on = []
  bucket     = "b"
  acl        = "private"
  tags = [
    "a",
    "b",
  ]
  id = "id"
  list_attr = [
    "x",
    "y",
  ]

  provisioner "local-exec" {
    command = "echo"
  }

  lifecycle {
    prevent_destroy = true
  }
}
