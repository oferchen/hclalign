data "aws_ami" "example" {
  provider    = aws.us
  count       = 1
  depends_on  = []
  owners      = ["amazon"]
  most_recent = true
  bar         = "bar"
  foo         = "foo"
}
