output "sentry_project_dsns" {
  description = "Sentry DSNs by project slug."

  value = {
    for slug, key in sentry_key.this :
    slug => key.dsn["public"]
  }

  sensitive = true
}