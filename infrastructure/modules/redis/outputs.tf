output "redis_url" {
  value     = "rediss://default:${upstash_redis_database.exampleDB.password}@${upstash_redis_database.exampleDB.endpoint}:${upstash_redis_database.exampleDB.port}"
  sensitive = true
}