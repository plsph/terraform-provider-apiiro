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

### Read-Only

- `id` (String)
- `name` (String)
- `provider` (String)
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
