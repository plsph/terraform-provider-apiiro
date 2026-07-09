data "apiiro_applicationgroups" "core" {
  name = "Core"
}

output "core_group_keys" {
  value = [for group in data.apiiro_applicationgroups.core.application_groups : group.key]
}
