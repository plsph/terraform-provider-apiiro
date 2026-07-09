---
page_title: "apiiro_teams Resource"
subcategory: "Organizational Teams"
description: |-
  Manages organizational teams in Apiiro.
---

# apiiro_teams Resource

Manages organizational teams in Apiiro.

## Example Usage

```terraform
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
```

## Import

```terraform
import {
  to = apiiro_teams.platform_engineering
  id = "<team-key>"
}
```
