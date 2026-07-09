---
page_title: "apiiro_engagements Data Source"
subcategory: "Engagements"
description: |-
  Use this data source to list engagements from Apiiro and optionally filter by name, type, status, or lead key.
---

# apiiro_engagements Data Source

Use this data source to retrieve engagements from Apiiro.

## Example Usage

```terraform
data "apiiro_engagements" "open_security_reviews" {
  type   = "SecurityReview"
  status = "Open"
}

output "open_security_review_names" {
  value = [for item in data.apiiro_engagements.open_security_reviews.engagements : item.name]
}
```

## Schema

### Optional

- `name` (String) Optional case-insensitive engagement name filter.
- `type` (String) Optional engagement type filter.
- `status` (String) Optional engagement status filter.
- `lead_key` (String) Optional engagement lead key filter.

### Read-Only

- `id` (String) Data source cache key.
- `engagements` (List of Object) Matching engagements.
