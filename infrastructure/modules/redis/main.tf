resource "upstash_redis_database" "this" {
  database_name  = var.redis_name
  region         = var.region
  primary_region = var.primary_region
  tls            = var.tls
}