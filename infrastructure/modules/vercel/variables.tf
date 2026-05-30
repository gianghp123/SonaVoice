variable "environment" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "name" {
  type = string
}

variable "default_regions" {
  type    = list(string)
  default = null
}

variable "framework" {
  type = string
}

variable "root_directory" {
  type = string
}

variable "environment_variables" {
  type = map(object({
    value     = string
    sensitive = optional(bool, false)
    target    = optional(list(string))
  }))

  default = {}
}

variable "target" {
  type    = list(string)
  default = ["production", "preview"]
}


variable "ignore_command" {
  type    = string
  default = null
}