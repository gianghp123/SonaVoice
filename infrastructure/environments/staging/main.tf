data "terraform_remote_state" "shared" {
  backend = "remote"

  config = {
    organization = "gianghp"

    workspaces = {
      name = "sona-shared"
    }
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
  source   = "../../modules/sentry"
  for_each = var.sentry_projects

  project             = var.app.project
  environment         = var.app.environment
  sentry_organization = var.sentry_credential.organization
  sentry_project = {
    slug        = each.key
    name        = each.value.name
    platform    = each.value.platform
    teams       = each.value.teams
    resolve_age = each.value.resolve_age
  }
}


locals {
  api_env_vars = merge(var.sona_go_api.environment_variables, {
    DATABASE_URL = {
      value     = module.neon_database.database_url
      sensitive = true
    }

    REDIS_URL = {
      value     = data.terraform_remote_state.shared.outputs.redis_url
      sensitive = true
    }

    SENTRY_DSN = {
      value     = module.sentry["sona-go-api"].dsn
      sensitive = true
    }
  })

  web_env_vars = merge(var.sona_nextjs.environment_variables, {
    API_URL = {
      value     = "${module.vercel_api.project_url}/api/v1"
      sensitive = false
    }

    SENTRY_DSN = {
      value     = module.sentry["sona-nextjs"].dsn
      sensitive = true
    }

    SENTRY_ORG = {
      value     = module.sentry["sona-nextjs"].organization
      sensitive = false
    }

    SENTRY_PROJECT = {
      value     = module.sentry["sona-nextjs"].project
      sensitive = false
    }
  })
}

module "vercel_api" {
  source = "../../modules/vercel"

  name            = "sona-go-api"
  environment     = var.app.environment
  github_repo     = var.github_repo
  framework       = var.sona_go_api.framework
  root_directory  = var.sona_go_api.root_directory
  default_regions = var.sona_go_api.default_regions

  environment_variables = local.api_env_vars
  target                = var.vercel_target

  ignore_command = coalesce(
    var.sona_go_api.ignore_command,
    "git diff HEAD^ HEAD --quiet -- ."
  )
}

module "vercel_web" {
  source = "../../modules/vercel"

  name           = "sona-nextjs"
  environment    = var.app.environment
  github_repo    = var.github_repo
  framework      = var.sona_nextjs.framework
  root_directory = var.sona_nextjs.root_directory

  environment_variables = local.web_env_vars
  target                = var.vercel_target

  ignore_command = coalesce(
    var.sona_nextjs.ignore_command,
    "git diff HEAD^ HEAD --quiet -- ."
  )
}