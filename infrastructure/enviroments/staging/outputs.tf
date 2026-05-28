output "database_url" {
  sensitive = true
  value = module.neon_database.database_url
}

output "sentry_project_dsns" {
  value     = module.sentry.sentry_project_dsns
  sensitive = true
}

output "website_url" {
  value = module.vercel_web.project_url
}

output "api_url" {
  value = module.vercel_api.project_url
  sensitive = true
}