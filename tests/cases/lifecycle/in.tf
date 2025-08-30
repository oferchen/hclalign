resource "null_resource" "example" {
  lifecycle {
    foo                   = "foo"
    create_before_destroy = true
    bar                   = "bar"
    prevent_destroy       = true
    ignore_changes        = []
    baz                   = "baz"
    replace_triggered_by  = []
  }
}
