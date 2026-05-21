output "database_url" {
  sensitive = true

  value = format(
    "postgresql://%s:%s@%s/%s?sslmode=require",
    neon_role.this.name,
    neon_role.this.password,
    neon_endpoint.this.host,
    neon_database.this.name
  )
}