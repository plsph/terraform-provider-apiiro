---
page_title: "apiiro_rolegroups Resource"
subcategory: "Role Groups"
description: |-
  Manages role groups in Apiiro.
---

# apiiro_rolegroups Resource

Manages role groups in Apiiro.

## Example Usage

```terraform
resource "apiiro_rolegroups" "security_admins" {
  name        = "Security Admins"
  description = "Security administration group"

  admin_ids = ["user-1", "user-2"]
}
```

## Import

```terraform
import {
  to = apiiro_rolegroups.security_admins
  id = "<role-group-key>"
}
```
