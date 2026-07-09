---
page_title: "apiiro_roles Resource"
subcategory: "Roles"
description: |-
  Manages roles in Apiiro.
---

# apiiro_roles Resource

Manages roles in Apiiro.

## Example Usage

```terraform
resource "apiiro_roles" "read_only_security" {
  name        = "Read Only Security"
  description = "Read-only role for security stakeholders"

  permissions = {
    Risks         = "Read"
    Applications  = "Read"
    Repositories  = "Read"
  }
}
```

## Import

```terraform
import {
  to = apiiro_roles.read_only_security
  id = "<role-key>"
}
```
