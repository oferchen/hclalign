data "aws_ami" "example" {
  owners      = ["amazon"]
  bar         = "bar"
  provider    = aws.us
  foo         = "foo"
  count       = 1
  depends_on  = []
  most_recent = true
}
