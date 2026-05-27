terraform {
  required_providers {
    upstash = {
      source  = "upstash/upstash"
      version = "2.1.0"
    }
  }
}

provider "upstash" {
  api_key = var.upstash_credential.api_key
  email   = var.upstash_credential.email
}
