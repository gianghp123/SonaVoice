resource "upstash_redis_database" "exampleDB" {
  database_name   = "${var.project}-${var.redis_name}-${var.environment}"
  region          = var.region
  primary_region  = var.primary_region
  tls             = var.tls
}