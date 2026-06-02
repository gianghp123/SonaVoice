output "dsn" {
  description = "Sentry DSN for the project."
  value       = sentry_key.this.dsn["public"]
  sensitive   = true
}

output "organization" {
  description = "Sentry organization slug."
  value       = var.sentry_organization
}

output "project" {
  description = "Sentry project slug."
  value       = sentry_project.this.slug
}
