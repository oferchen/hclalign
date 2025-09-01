output "example" {
  # doc1
  // doc2
  /* doc3 */
  description = "desc" // desc trailing
  value = 1


  /* before sensitive */
  sensitive = true // sens trailing

  // mid
  depends_on = [var.x] /* dep trailing */
}
