resource "vercel_project" "this" {
  name           = "${var.name}-${var.environment}"
  framework      = var.framework
  root_directory = var.root_directory

  git_repository = {
    type = "github"
    repo = var.github_repo
  }

  environment = [
    for env_key, env_value in var.environment_variables : {
      key       = env_key
      value     = env_value.value
      sensitive = env_value.sensitive
      target    = coalesce(env_value.target, var.target)
    }
  ]
}
