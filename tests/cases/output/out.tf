output "example" {
  description = "desc"
  value       = var.a
  sensitive   = true
}

output "unknown" {
  description = "other"
  value       = var.b
  foo         = "bar"
  baz         = "qux"
}

output "depends" {
  value      = var.c
  sensitive  = false
  ephemeral  = true
  depends_on = [var.x]
  foo        = "bar"
}

output "already" {
  description = "foo"
  value       = var.foo
  sensitive   = false
  depends_on  = [var.dep]
}
