terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.13"
    }
    sentry = {
      source = "jianyuan/sentry"
    }
    vercel = {
      source  = "vercel/vercel"
      version = ">= 4.8"
    }
  }
  cloud {
    organization = "gianghp"

    workspaces {
      name = "sona-staging"
    }
  }
}

provider "neon" {
  api_key = var.neon_api_key
}

provider "sentry" {
  token = var.sentry_credential.api_key
}

provider "vercel" {
  api_token = var.vercel_api_token
}