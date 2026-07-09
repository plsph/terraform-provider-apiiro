resource "apiiro_applicationgroups" "core_platform" {
  name         = "Core Platform"
  applications = ["app-key-1", "app-key-2"]
  tags         = ["Critical", "Tier1"]
}
