data "apiiro_connectors" "github" {
  provider_name = "Github"
}

output "github_connector_urls" {
  value = [for connector in data.apiiro_connectors.github.connectors : connector.url]
}
