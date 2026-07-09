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
