data "apiiro_auditlogs" "failed_events" {
  status = "Failure"
}

output "failed_audit_event_ids" {
  value = [for log in data.apiiro_auditlogs.failed_events.logs : log.id]
}
