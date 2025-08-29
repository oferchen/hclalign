data "aws_ami" "example" {
  depends_on  = []
  count       = 1
  provider    = aws.us
  owners      = ["amazon"]
  most_recent = true
  bar         = "bar"
  foo         = "foo"
}
