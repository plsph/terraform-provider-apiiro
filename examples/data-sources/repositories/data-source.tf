data "apiiro_repositories" "all" {}

output "repository_names" {
  value = [for repo in data.apiiro_repositories.all.repositories : repo.name]
}
