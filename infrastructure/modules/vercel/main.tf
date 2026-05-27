resource "vercel_project" "this" {
  for_each = var.vercel_projects

  name           = "${each.key}-${var.environment}"
  framework      = each.value.framework
  root_directory = each.value.root_directory

  git_repository = {
    type = "github"
    repo = var.github_repo
  }

  environment = [
    for env_key, env_value in each.value.environment_variables : {
      key       = env_key
      value     = env_value.value
      sensitive = env_value.sensitive
      target    = var.target
    }
  ]
}

