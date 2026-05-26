resource "neon_branch" "this" {
  project_id = var.neon_project_id
  name       = var.branch
}

resource "neon_endpoint" "this" {
  project_id = var.neon_project_id
  branch_id  = neon_branch.this.id

  autoscaling_limit_min_cu = 0.25
  autoscaling_limit_max_cu = 1
}

resource "neon_role" "this" {
  project_id = var.neon_project_id
  branch_id  = neon_branch.this.id
  name       = var.role_name
  depends_on = [neon_endpoint.this]
}

resource "neon_database" "this" {
  project_id = var.neon_project_id
  branch_id  = neon_branch.this.id
  owner_name = neon_role.this.name
  name       = var.database_name
  depends_on = [
    neon_endpoint.this,
    neon_role.this
  ]
}