---
page_title: "Apiiro Provider"
subcategory: ""
description: |-
  The Apiiro provider lets you manage Apiiro resources using Terraform.
---

# Apiiro Provider

Use the Apiiro provider to configure and manage security-relevant configuration in Apiiro.

## Example Usage

```terraform
terraform {
  required_providers {
    apiiro = {
      source  = "plsph/apiiro"
      version = "0.1.0"
    }
  }
}

provider "apiiro" {
  base_url = "https://app.apiiro.com"
  token    = var.apiiro_token
}
```

## Schema

### Optional

- `base_url` (String) Apiiro API base URL. Defaults to `https://app.apiiro.com` or environment variable `APIIRO_BASE_URL`.
- `token` (String, Sensitive) Apiiro access token. Can also be set via environment variable `APIIRO_TOKEN`.
