resource "null_resource" "example" {
  connection {
    user    = "ubuntu"
    host    = "example.com"
    timeout = "1m"
  }

  provisioner "local-exec" {
    connection {
      timeout = "1m"
      user    = "root"
      host    = "127.0.0.1"
    }
  }
}
