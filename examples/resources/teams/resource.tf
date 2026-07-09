resource "apiiro_teams" "platform_engineering" {
  name        = "Platform Engineering"
  description = "Team owning platform and shared services"

  project_urls = [
    "https://github.com/acme"
  ]

  tags = {
    Department = "Engineering"
  }
}
