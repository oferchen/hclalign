resource "null_resource" "example" {

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = []
    replace_triggered_by  = []
    foo                   = "foo"
    bar                   = "bar"
    baz                   = "baz"
  }
}
