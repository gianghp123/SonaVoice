variable "app" {
  type = object({
    project     = string
    region      = string
    environment = string
  })
}

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

variable "vercel_api_token" {
  type      = string
  sensitive = true
}

variable "github_repo" {
  type = string
}

variable "sona_nextjs" {
  type = object({
    framework      = string
    root_directory = string

    environment_variables = optional(map(object({
      value     = string
      sensitive = optional(bool, false)
    })), {})
  })
}

variable "sona_go_api" {
  type = object({
    framework      = string
    root_directory = string

    environment_variables = optional(map(object({
      value     = string
      sensitive = optional(bool, false)
    })), {})
  })
}

variable "vercel_target" {
  type    = list(string)
  default = ["production", "preview", "development"]
}
