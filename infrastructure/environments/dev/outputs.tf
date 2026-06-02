output "database_url" {
  sensitive = true
  value     = module.neon_database.database_url
}

output "sentry_project_dsns" {
  value = {
    for k, v in module.sentry : k => v.dsn
  }
  sensitive = true
}
