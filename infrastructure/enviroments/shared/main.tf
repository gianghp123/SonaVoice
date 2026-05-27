module "upstash_redis" {
  source         = "../../modules/redis"
  redis_name     = var.upstash_redis_config.name
  region         = var.upstash_redis_config.region
  primary_region = var.upstash_redis_config.primary_region
  tls            = var.upstash_redis_config.tls
}