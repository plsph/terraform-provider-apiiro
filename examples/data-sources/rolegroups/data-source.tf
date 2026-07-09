data "apiiro_rolegroups" "all" {}

output "role_group_keys" {
  value = [for group in data.apiiro_rolegroups.all.groups : group.key]
}
