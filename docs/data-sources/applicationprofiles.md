---
page_title: "apiiro_applicationprofiles Data Source"
subcategory: "Application Profiles"
description: |-
  Use this data source to list application profiles from Apiiro and optionally filter by key, name, type, risk level, or internet exposure.
---

# apiiro_applicationprofiles Data Source

Use this data source to retrieve read-only application profile metadata from Apiiro.

## Example Usage

```terraform
data "apiiro_applicationprofiles" "internet_exposed_high" {
  internet_exposed = true
  risk_level       = "High"
}

output "high_risk_public_apps" {
  value = [for profile in data.apiiro_applicationprofiles.internet_exposed_high.application_profiles : profile.name]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive name filter.
- `application_key` (String) Optional exact application key filter.
- `application_type` (String) Optional application type filter.
- `risk_level` (String) Optional risk level filter.
- `internet_exposed` (Boolean) Optional internet exposure filter.

### Read-Only

- `id` (String) Data source cache key.
- `application_profiles` (List of Object) Matching application profiles.
