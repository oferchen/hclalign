resource "null_resource" "example" {

  provisioner "local-exec" {
    when       = "destroy"
    on_failure = "continue"
    bar        = "b"

    connection {
      host = "example.com"
    }
    foo = "f"
  }
}
