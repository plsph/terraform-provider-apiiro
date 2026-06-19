---
page_title: "apiiro_scmrepositories Data Source"
subcategory: "SCM Repositories"
description: |-
  Use this data source to list SCM repositories from Apiiro and optionally filter by repository name or repository key.
---

# apiiro_scmrepositories Data Source

Use this data source to retrieve SCM repositories from Apiiro.

## Example Usage

```terraform
data "apiiro_scmrepositories" "all" {}

output "repository_names" {
  value = [for repo in data.apiiro_scmrepositories.all.repositories : repo.name]
}
```

```terraform
data "apiiro_scmrepositories" "filtered" {
  repository_name = "payment-service"
}

output "first_repository_key" {
  value = try(data.apiiro_scmrepositories.filtered.repositories[0].key, null)
}
```

## Schema

### Optional

- `repository_name` (String) Optional case-insensitive filter for SCM repository name.
- `scm_repository_key` (String) Optional exact filter for SCM repository key.

### Read-Only

- `id` (String) Data source cache key.
- `repositories` (List of Object) Matching SCM repositories.

### Nested Schema for `repositories`

Read-only attributes:

- `key` (String)
- `id` (String)
- `name` (String)
- `provider` (String)
- `server_url` (String)
- `url` (String)
- `project_url` (String)
- `http_clone_url` (String)
- `ssh_url` (String)
- `project_id` (String)
- `project_external_id` (String)
- `repository_external_id` (String)
- `default_branch` (String)
- `monitor_status` (String)
- `monitored_branches` (List of String)
- `languages` (List of String)
- `associated_project_ids` (List of String)
- `approximate_size_mb` (Number)
- `approximate_file_count` (Number)
- `is_public` (Boolean)
- `is_archived` (Boolean)
- `is_ignored` (Boolean)
- `ignore_reason` (String)
- `ignored_by` (String)
- `created_at` (String)
- `last_monitoring_change_timestamp` (String)
