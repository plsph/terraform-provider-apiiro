---
page_title: "apiiro_applications Data Source"
subcategory: "Applications"
description: |-
  Use this data source to list applications from Apiiro and optionally filter by name, key, or application type.
---

# apiiro_applications Data Source

Use this data source to retrieve application configurations from Apiiro.

## Example Usage

```terraform
data "apiiro_applications" "all" {}

output "application_names" {
  value = [for app in data.apiiro_applications.all.applications : app.name]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive name filter.
- `application_key` (String) Optional exact application key filter.
- `application_type` (String) Optional application type filter.

### Read-Only

- `id` (String) Data source cache key.
- `applications` (List of Object) Matching applications.
