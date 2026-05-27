data "terraform_remote_state" "shared" {
  backend = "local"

  config = {
    path = "../shared/terraform.tfstate"
  }
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


locals {
  env_vars = {
    DATABASE_URL = {
      value     = module.neon_database.database_url
      sensitive = true
    }
    REDIS_URL = {
      value     = data.terraform_remote_state.shared.outputs.redis_url
      sensitive = true
    }
    SENTRY_PROJECT_DSNS = {
      value     = module.sentry.sentry_project_dsns
      sensitive = true
    }
  }

  vercel_projects = {
    for name, cfg in var.vercel_projects : name => {
      framework      = cfg.framework
      root_directory = cfg.root_directory
      environment_variables = {
        for k, v in cfg.environment_variables : k => (
          v != null ? v :
          k == "SENTRY_DSN" ? { value = module.sentry.sentry_project_dsns[name], sensitive = true } :
          local.env_vars[k]
        )
      }
    }
  }

  missing_dynamic_keys = flatten([
    for name, cfg in var.vercel_projects : [
      for k, v in cfg.environment_variables : "${name}.${k}"
      if v == null && k != "SENTRY_DSN" && !contains(keys(local.env_vars), k)
    ]
  ])
}

check "dynamic_values_exist" {
  assert {
    condition     = length(local.missing_dynamic_keys) == 0
    error_message = "Missing dynamic values: ${jsonencode(local.missing_dynamic_keys)}"
  }
}

check "dynamic_values_not_null" {
  assert {
    condition = alltrue([
      for name, proj in local.vercel_projects :
      alltrue([for k, v in proj.environment_variables : v != null])
    ])
    error_message = "Some dynamic values resolved to null (module output missing)"
  }
}

module "vercel" {
  source          = "../../modules/vercel"
  vercel_projects = local.vercel_projects
  github_repo     = var.github_repo
  project         = var.app.project
  environment     = var.app.environment
  target          = var.vercel_target
}
