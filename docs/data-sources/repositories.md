---
page_title: "apiiro_repositories Data Source"
subcategory: "Repositories"
description: |-
  Use this data source to list repositories from Apiiro and optionally filter by repository name, repository key, SCM repository key, or provider.
---

# apiiro_repositories Data Source

Use this data source to retrieve repositories from Apiiro v2.

## Example Usage

```terraform
data "apiiro_repositories" "all" {}

output "repository_names" {
  value = [for repo in data.apiiro_repositories.all.repositories : repo.name]
}
```

```terraform
data "apiiro_repositories" "filtered" {
  name = "payment-service"
}

output "first_repository_key" {
  value = try(data.apiiro_repositories.filtered.repositories[0].key, null)
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive filter for repository name.
- `key` (String) Optional exact filter for repository key.
- `scm_repository_key` (String) Optional exact filter for SCM repository key.
- `provider_name` (String) Optional exact filter for provider.

### Read-Only

- `id` (String) Data source cache key.
- `repositories` (List of Object) Matching repositories.

### Nested Schema for `repositories`

Read-only attributes:

- `active_since` (String)
- `api_count` (Number)
- `branch_name` (String)
- `business_impact` (String)
- `contributor_count` (Number)
- `dependency_count` (Number)
- `external_id` (String)
- `has_data_models` (Boolean)
- `has_external_dependencies` (Boolean)
- `has_payments_data` (Boolean)
- `has_phi_data` (Boolean)
- `has_pii_data` (Boolean)
- `has_sensitive_dependencies` (Boolean)
- `insights` (List of String)
- `is_active` (Boolean)
- `is_archived` (Boolean)
- `is_default_branch` (Boolean)
- `is_deployed` (Boolean)
- `is_internet_exposed` (Boolean)
- `is_public` (Boolean)
- `key` (String)
- `language_percentages` (Map of Number)
- `languages` (List of String)
- `last_activity` (String)
- `licenses` (List of String)
- `modules` (List of Object)
- `name` (String)
- `project_id` (String)
- `provider` (String)
- `risk_level` (String)
- `risk_score` (Number)
- `scm_repository_key` (String)
- `server_url` (String)
- `url` (String)

### Nested Schema for `modules`

- `file_path` (String)
- `name` (String)
