output "demo" {
  ephemeral  = true
  value      = var.v
  depends_on = [var.dep]
  description = "desc"
  sensitive  = true
}
