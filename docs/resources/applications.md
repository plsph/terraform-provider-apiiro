---
page_title: "apiiro_applications Resource"
subcategory: "Applications"
description: |-
  Manages applications in Apiiro.
---

# apiiro_applications Resource

Manages application configuration in Apiiro.

## Example Usage

```terraform
resource "apiiro_applications" "payment_gateway" {
  name               = "Payment Gateway"
  application_type   = "Service"
  business_impact    = "High"
  is_internet_facing = true

  repository_urls = [
    "https://github.com/acme/payment-service"
  ]

  tags = {
    Team       = "Platform"
    Compliance = "SOC2"
  }
}
```

## Import

```terraform
import {
  to = apiiro_applications.payment_gateway
  id = "<application-key>"
}
```
