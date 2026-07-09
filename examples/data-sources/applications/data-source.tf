data "apiiro_applications" "service_apps" {
  application_type = "Service"
}

output "service_application_keys" {
  value = [for app in data.apiiro_applications.service_apps.applications : app.key]
}
