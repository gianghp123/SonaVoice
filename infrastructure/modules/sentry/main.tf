resource "sentry_project" "this" {
  for_each = var.sentry_projects

  organization = var.sentry_organization

  name     = "${each.value.name}-${var.environment}"
  slug     = each.key
  platform = each.value.platform
  teams    = each.value.teams

  resolve_age = each.value.resolve_age
}

resource "sentry_key" "this" {
  for_each = sentry_project.this

  organization = var.sentry_organization
  project      = each.value.slug
  name         = "${each.value.name} Client Key"
}