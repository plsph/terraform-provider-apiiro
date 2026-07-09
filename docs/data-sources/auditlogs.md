---
page_title: "apiiro_auditlogs Data Source"
subcategory: "Audit Logs"
description: |-
  Use this data source to list audit logs from Apiiro and optionally filter by event type, status, user, or user type.
---

# apiiro_auditlogs Data Source

Use this data source to retrieve audit log records from Apiiro.

## Example Usage

```terraform
data "apiiro_auditlogs" "failed_events" {
  status = "Failure"
}

output "failed_audit_event_ids" {
  value = [for log in data.apiiro_auditlogs.failed_events.logs : log.id]
}
```

## Schema

### Optional

- `event_type` (String) Optional exact event type filter.
- `status` (String) Optional status filter (`Success` or `Failure`).
- `user` (String) Optional case-insensitive substring user filter.
- `user_type` (String) Optional exact user type filter.

### Read-Only

- `id` (String) Data source cache key.
- `logs` (List of Object) Matching audit log entries.
