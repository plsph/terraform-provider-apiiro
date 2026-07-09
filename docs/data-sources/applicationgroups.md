---
page_title: "apiiro_applicationgroups Data Source"
subcategory: "Application Groups"
description: |-
  Use this data source to list application groups from Apiiro and optionally filter by group name or key.
---

# apiiro_applicationgroups Data Source

Use this data source to retrieve application groups from Apiiro.

## Example Usage

```terraform
data "apiiro_applicationgroups" "all" {}

output "application_group_keys" {
  value = [for group in data.apiiro_applicationgroups.all.application_groups : group.key]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive name filter.
- `key` (String) Optional exact group key filter.

### Read-Only

- `id` (String) Data source cache key.
- `application_groups` (List of Object) Matching application groups.
