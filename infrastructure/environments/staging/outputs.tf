output "database_url" {
  sensitive = true
  value     = module.neon_database.database_url
}

output "sentry_project_dsns" {
  value     = {
    for k, v in module.sentry : k => v.dsn
  }
  sensitive = true
}

output "website_url" {
  value = module.vercel_web.project_url
}

output "api_url" {
  value = module.vercel_api.project_url
}

output "vercel_api_project_name" {
  value = module.vercel_api.project_name
}