output "database_url" {
  sensitive = true
  value     = module.neon_database.database_url
}

output "sentry_project_dsns" {
  value     = module.sentry.sentry_project_dsns
  sensitive = true
}
