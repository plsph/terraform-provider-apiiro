terraform {
  required_providers {
    apiiro = {
      source  = "plsph/apiiro"
      version = "0.1.0"
    }
  }
}

provider "apiiro" {
  base_url = "https://app.apiiro.com"
  token    = var.apiiro_token
}

variable "apiiro_token" {
  type      = string
  sensitive = true
}
