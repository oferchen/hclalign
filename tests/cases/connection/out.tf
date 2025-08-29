resource "null_resource" "example" {

  connection {
    host    = "example.com"
    timeout = "1m"
    user    = "ubuntu"
  }

  provisioner "local-exec" {

    connection {
      host    = "127.0.0.1"
      timeout = "1m"
      user    = "root"
    }
  }
}
