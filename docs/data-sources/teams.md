---
page_title: "apiiro_teams Data Source"
subcategory: "Organizational Teams"
description: |-
  Use this data source to list organizational teams from Apiiro and optionally filter by team name or key.
---

# apiiro_teams Data Source

Use this data source to retrieve organizational teams from Apiiro.

## Example Usage

```terraform
data "apiiro_teams" "platform" {
  name = "Platform"
}

output "platform_team_keys" {
  value = [for team in data.apiiro_teams.platform.teams : team.key]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive name filter.
- `key` (String) Optional exact team key filter.

### Read-Only

- `id` (String) Data source cache key.
- `teams` (List of Object) Matching teams.
