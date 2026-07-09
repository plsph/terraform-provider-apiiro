---
page_title: "apiiro_roles Data Source"
subcategory: "Roles"
description: |-
  Use this data source to list roles from Apiiro and optionally filter by role name or key.
---

# apiiro_roles Data Source

Use this data source to retrieve roles from Apiiro.

## Example Usage

```terraform
data "apiiro_roles" "engineering" {
  name = "Engineering"
}

output "engineering_roles" {
  value = [for role in data.apiiro_roles.engineering.roles : role.name]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive role name filter.
- `key` (String) Optional exact role key filter.

### Read-Only

- `id` (String) Data source cache key.
- `roles` (List of Object) Matching roles.
