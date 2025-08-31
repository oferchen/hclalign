output "demo" {
  description = "desc"
  value       = var.v
  sensitive   = true
  ephemeral   = true
  depends_on  = [var.dep]
}
