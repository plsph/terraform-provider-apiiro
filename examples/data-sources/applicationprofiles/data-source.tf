data "apiiro_applicationprofiles" "internet_exposed_high" {
  internet_exposed = true
  risk_level       = "High"
}

output "internet_exposed_high_risk_profiles" {
  value = [for profile in data.apiiro_applicationprofiles.internet_exposed_high.application_profiles : profile.name]
}
