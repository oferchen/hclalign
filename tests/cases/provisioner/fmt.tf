resource "null_resource" "example" {
  provisioner "local-exec" {
    bar  = "b"
    when = "destroy"
    connection {
      host = "example.com"
    }
    foo        = "f"
    on_failure = "continue"
  }
}
