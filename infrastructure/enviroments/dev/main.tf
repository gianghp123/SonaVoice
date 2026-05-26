terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.13"
    }
    sentry = {
      source = "jianyuan/sentry"
    }
  }
}

provider "neon" {
  api_key = var.neon_api_key
}

provider "sentry" {
  token = var.sentry_api_key
}

module "database" {
  source = "../../modules/database"
  project = var.project
  region = var.region
  environment = var.environment
  branch = var.branch
  database_name = var.database_name
  role_name = var.role_name
  neon_project_id = var.neon_project_id
}


module "sentry" {
  source = "../../modules/sentry"
  project = var.project
  environment = var.environment
  sentry_organization = var.sentry_organization
  sentry_projects = var.sentry_projects
}