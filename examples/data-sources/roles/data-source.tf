data "apiiro_roles" "security" {
  name = "Security"
}

output "security_role_keys" {
  value = [for role in data.apiiro_roles.security.roles : role.key]
}
