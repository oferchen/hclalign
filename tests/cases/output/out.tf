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
  foo        = "bar"
  depends_on = [var.x]
}

output "already" {
  description = "foo"
  value       = var.foo
  sensitive   = false
}
