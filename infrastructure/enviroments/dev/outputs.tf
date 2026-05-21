output "database_url" {
  sensitive = true
  value = module.database.database_url
}