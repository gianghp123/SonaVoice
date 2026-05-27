output "redis_url" {
  value = module.upstash_redis.redis_url
  sensitive = true
}