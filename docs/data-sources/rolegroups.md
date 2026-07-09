---
page_title: "apiiro_rolegroups Data Source"
subcategory: "Role Groups"
description: |-
  Use this data source to list role groups from Apiiro and optionally filter by group name or key.
---

# apiiro_rolegroups Data Source

Use this data source to retrieve role groups from Apiiro.

## Example Usage

```terraform
data "apiiro_rolegroups" "all" {}

output "role_group_names" {
  value = [for group in data.apiiro_rolegroups.all.groups : group.name]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive group name filter.
- `key` (String) Optional exact group key filter.

### Read-Only

- `id` (String) Data source cache key.
- `groups` (List of Object) Matching role groups.
