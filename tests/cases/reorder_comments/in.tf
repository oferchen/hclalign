output "example" {
  value = 1


  # doc1
  // doc2
  /* doc3 */
  description = "desc" // desc trailing

  // mid
  depends_on = [var.x] /* dep trailing */


  /* before sensitive */
  sensitive = true // sens trailing
}
