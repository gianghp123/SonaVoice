variable "project" {
  type = string
}

variable "environment" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "vercel_projects" {
  type = map(object({
    framework      = string
    root_directory = string

    environment_variables = map(object({
      value     = string
      sensitive = optional(bool, false)
    }))
  }))
}

variable "target" {
  type = list(string)
}