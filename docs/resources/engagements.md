---
page_title: "apiiro_engagements Resource"
subcategory: "Engagements"
description: |-
  Manages engagements in Apiiro.
---

# apiiro_engagements Resource

Manages engagements in Apiiro.

## Example Usage

```terraform
resource "apiiro_engagements" "q3_security_review" {
  name       = "Q3 Security Review"
  type       = "SecurityReview"
  start_date = "2026-07-01"
  status     = "Open"
  summary    = "Quarterly security review for internet-facing applications"

  scope = {
    application_keys = ["app-key-1", "app-key-2"]
  }

  tags = {
    Quarter = "Q3"
    Year    = "2026"
  }
}
```

## Import

```terraform
import {
  to = apiiro_engagements.q3_security_review
  id = "<engagement-key>"
}
```
