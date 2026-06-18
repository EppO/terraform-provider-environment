# Provider-defined functions require Terraform 1.8 or later.
terraform {
  required_version = ">= 1.8.0"

  required_providers {
    environment = {
      source = "EppO/environment"
    }
  }
}

# Return every environment variable as a map.
output "all" {
  value = provider::environment::get_env_map("")
}

# Filter variable names with a regular expression.
output "lc" {
  value = provider::environment::get_env_map("^LC_")
}

# Wrap the result with sensitive() when it may contain secrets.
output "tokens" {
  value     = provider::environment::get_env_map("TOKEN")
  sensitive = true
}
