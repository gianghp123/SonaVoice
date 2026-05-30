variable "project" {
  type = string
}

variable "neon_project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type    = string
  default = "development"
}

variable "database_name" {
  type = string
}

variable "role_name" {
  type = string
}
