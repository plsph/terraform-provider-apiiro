data "apiiro_engagements" "open_security_reviews" {
  type   = "SecurityReview"
  status = "Open"
}

output "open_security_review_keys" {
  value = [for engagement in data.apiiro_engagements.open_security_reviews.engagements : engagement.key]
}
