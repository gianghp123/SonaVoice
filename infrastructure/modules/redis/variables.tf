variable "redis_name" {
  type = string
}

variable "region" {
  type = string
}

variable "primary_region" {
  type = string
}

variable "tls" {
  type = bool
  default = true
}

variable "project" {
  type = string
}

variable "environment" {
  type = string
}