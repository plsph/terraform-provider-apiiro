---
page_title: "apiiro_scmrepositories Resource"
subcategory: "SCM Repositories"
description: |-
  Manages monitoring configuration and tags for an existing SCM repository in Apiiro.
---

# apiiro_scmrepositories Resource

Manages monitoring state, monitored branches, and tags for an existing Apiiro SCM repository.

## Example Usage

```terraform
resource "apiiro_scmrepositories" "payment_service" {
  scm_repository_key = "https_github_com_acme_payment-service"
  monitored          = true
  monitored_branches = ["main", "release"]

  tags = {
    Compliance = "SOC2"
    Team       = "Platform"
  }
}
```

## Import

Import an SCM repository resource by repository key:

```terraform
import {
  to = apiiro_scmrepositories.payment_service
  id = "https_github_com_acme_payment-service"
}
```

## Schema

### Required

- `scm_repository_key` (String) Unique Apiiro SCM repository key.

### Optional

- `monitored` (Boolean) Whether the repository should be monitored.
- `monitored_branches` (Set of String) Specific monitored branches.
- `tags` (Map of String) Repository tags managed through Apiiro tag API.

## Tags behavior

- **Default (recommended)**: If you omit the `tags` attribute entirely, the provider treats repository tags as a computed field and will not surface or attempt to manage any tags that exist only on the backend (for example, internal/default tags such as `repository_type`). This prevents "provider produced inconsistent result after apply" errors caused by backend-created tags appearing after apply.
- **Manage specific tags**: To manage one or more tags, include the `tags` map with the keys you want to manage. The provider will ensure those keys exist with the given values; backend-only tags that are not in this map remain ignored.
- **Explicit empty map**: If you set `tags = {}` (an explicit empty map), the provider treats tags as managed and will keep the tag set empty in Terraform state. This will cause previously managed tags to be deleted from the repository.

## Examples (tag semantics)

Omit tags (do not manage tags — computed):

```terraform
resource "apiiro_scmrepositories" "no_tags" {
  scm_repository_key = "https_github_com_acme_payment-service"
  monitored          = true
}
```

Manage selected tags (only these keys are managed):

```terraform
resource "apiiro_scmrepositories" "selected_tags" {
  scm_repository_key = "https_github_com_acme_payment-service"
  tags = {
    Compliance = "SOC2"
  }
}
```

Manage tags as empty (explicitly remove previously-managed tags):

```terraform
resource "apiiro_scmrepositories" "empty_tags" {
  scm_repository_key = "https_github_com_acme_payment-service"
  tags = {}
}
```

### Read-Only

- `id` (String)
- `name` (String)
- `provider_name` (String)
- `monitor_status` (String)
- `default_branch` (String)
- `url` (String)
- `server_url` (String)
- `http_clone_url` (String)
- `ssh_url` (String)
- `repository_external_id` (String)
- `project_id` (String)
- `project_external_id` (String)
- `project_url` (String)
- `created_at` (String)
- `last_monitoring_change_timestamp` (String)
- `is_public` (Boolean)
- `is_archived` (Boolean)
- `is_ignored` (Boolean)
- `ignore_reason` (String)
- `ignored_by` (String)
