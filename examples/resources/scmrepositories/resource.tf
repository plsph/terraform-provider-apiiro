resource "apiiro_scmrepositories" "payment_service" {
  scm_repository_key = "https_github_com_acme_payment-service"
  monitored          = true
  monitored_branches = ["main", "release"]

  tags = {
    Compliance = "SOC2"
    Team       = "Platform"
  }
}
