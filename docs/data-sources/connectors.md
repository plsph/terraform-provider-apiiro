---
page_title: "apiiro_connectors Data Source"
subcategory: "Connectors"
description: |-
  Use this data source to list connectors from Apiiro and optionally filter by connector ID or provider name.
---

# apiiro_connectors Data Source

Use this data source to retrieve connectors from Apiiro.

## Example Usage

```terraform
data "apiiro_connectors" "github" {
  provider_name = "Github"
}

output "github_connector_urls" {
  value = [for connector in data.apiiro_connectors.github.connectors : connector.url]
}
```

## Schema

### Optional

- `connector_id` (String) Optional exact connector ID filter.
- `provider_name` (String) Optional exact provider filter.

### Read-Only

- `id` (String) Data source cache key.
- `connectors` (List of Object) Matching connectors.
