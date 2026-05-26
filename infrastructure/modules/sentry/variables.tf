variable "sentry_organization" {
  type = string
}

variable "sentry_projects" {
  type = map(object({
    name     = string
    platform = string
    teams    = list(string)

    resolve_age = optional(number, 720)
  }))
}

variable "environment" {
  type = string
}

variable "project" {
  type = string
}