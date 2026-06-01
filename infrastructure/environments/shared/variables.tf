variable "app" {
  type = object({
    region = string
  })
}

variable "upstash_credential" {
  type = object({
    email   = string
    api_key = string
  })

  sensitive = true
}

variable "upstash_redis_config" {
  type = object({
    name           = string
    region         = string
    primary_region = string
    tls            = bool
  })
}


variable "terraform_api_token" {
  type = string
}
