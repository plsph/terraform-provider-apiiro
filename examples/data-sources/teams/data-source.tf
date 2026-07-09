data "apiiro_teams" "platform" {
  name = "Platform"
}

output "platform_team_keys" {
  value = [for team in data.apiiro_teams.platform.teams : team.key]
}
