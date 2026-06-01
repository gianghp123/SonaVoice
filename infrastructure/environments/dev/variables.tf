variable "neon_config" {
  type = object({
    project_id    = string
    database_name = string
    role_name     = string
  })

  sensitive = true
}

variable "neon_api_key" {
  type      = string
  sensitive = true
}

variable "sentry_projects" {
  type = map(object({
    name        = string
    platform    = string
    teams       = list(string)
    resolve_age = optional(number, 720)
  }))
}

variable "sentry_credential" {
  type = object({
    api_key      = string
    organization = string
  })

  sensitive = true
}

variable "app" {
  type = object({
    project     = string
    region      = string
    environment = string
  })
}
