output "example" {
  value       = var.a
  sensitive   = true
  description = "desc"
}

output "unknown" {
  foo         = "bar"
  description = "other"
  value       = var.b
  baz         = "qux"
}

output "depends" {
  foo        = "bar"
  value      = var.c
  sensitive  = false
  ephemeral  = true
  depends_on = [var.x]
}

output "already" {
  description = "foo"
  value       = var.foo
  sensitive   = false
  depends_on  = [var.dep]
}
