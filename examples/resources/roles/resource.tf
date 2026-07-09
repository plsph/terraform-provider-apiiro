resource "apiiro_roles" "read_only_security" {
  name        = "Read Only Security"
  description = "Read-only role for security stakeholders"

  permissions = {
    Risks        = "Read"
    Applications = "Read"
    Repositories = "Read"
  }
}
