data "apiiro_scmrepositories" "all" {}

output "all_scm_repository_names" {
  value = [for repo in data.apiiro_scmrepositories.all.repositories : repo.name]
}

data "apiiro_scmrepositories" "payment_service" {
  repository_name = "payment-service"
}

output "payment_service_repository_key" {
  value = try(data.apiiro_scmrepositories.payment_service.repositories[0].key, null)
}
