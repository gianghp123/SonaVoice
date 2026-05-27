output "project_name" {
  value = vercel_project.this.name
}
output "project_url" {
  value = "https://${vercel_project.this.name}.vercel.app"
}

output "project_id" {
  value = vercel_project.this.id
}