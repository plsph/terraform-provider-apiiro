---
page_title: "apiiro_applicationgroups Resource"
subcategory: "Application Groups"
description: |-
  Manages application groups in Apiiro.
---

# apiiro_applicationgroups Resource

Manages application groups in Apiiro.

## Example Usage

```terraform
resource "apiiro_applicationgroups" "core_platform" {
  name         = "Core Platform"
  applications = ["app-key-1", "app-key-2"]
  tags         = ["Critical", "Tier1"]
}
```

## Import

```terraform
import {
  to = apiiro_applicationgroups.core_platform
  id = "<application-group-key>"
}
```
