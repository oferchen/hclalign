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
  depends_on  = [var.x]
  value       = var.c
}

output "already" {
  description = "foo"
  value       = var.foo
  sensitive   = false
}
