resource "sentry_project" "this" {
  organization = var.sentry_organization

  name     = "${var.sentry_project.name}-${var.environment}"
  slug     = "${var.sentry_project.slug}-${var.environment}"
  platform = var.sentry_project.platform
  teams    = var.sentry_project.teams

  resolve_age = var.sentry_project.resolve_age
}

resource "sentry_key" "this" {
  organization = var.sentry_organization
  project      = sentry_project.this.slug
  name         = "${sentry_project.this.name} Client Key"
}
