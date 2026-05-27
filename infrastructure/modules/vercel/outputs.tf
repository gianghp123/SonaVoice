output "vercel_project_urls" {
  value = {
    for name, proj in vercel_project.this : name => "https://${proj.name}.vercel.app"
  }
  description = "Vercel project URLs"
}