resource "apiiro_rolegroups" "security_admins" {
  name        = "Security Admins"
  description = "Security administration group"

  admin_ids = ["user-1", "user-2"]
}
