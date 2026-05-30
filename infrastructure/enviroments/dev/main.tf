terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.13"
    }
    sentry = {
      source = "jianyuan/sentry"
    }
    upstash = {
      source  = "upstash/upstash"
      version = "2.1.0"
    }
  }
}

provider "neon" {
  api_key = var.neon_api_key
}

provider "sentry" {
  token = var.sentry_credential.api_key
}

provider "upstash" {
  api_key = var.upstash_credential.api_key
  email   = var.upstash_credential.email
}

module "neon_database" {
  source = "../../modules/database"

  project     = var.app.project
  region      = var.app.region
  environment = var.app.environment

  database_name   = var.neon_config.database_name
  role_name       = var.neon_config.role_name
  neon_project_id = var.neon_config.project_id
}

module "sentry" {
  source = "../../modules/sentry"

  project             = var.app.project
  environment         = var.app.environment
  sentry_organization = var.sentry_credential.organization
  sentry_projects     = var.sentry_projects
}

data "terraform_remote_state" "shared" {
  backend = "local"

  config = {
    path = "../shared/terraform.tfstate"
  }
}