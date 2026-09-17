# Provider-defined functions require Terraform 1.8 or later.
terraform {
  required_version = ">= 1.8.0"

  required_providers {
    environment = {
      source = "EppO/environment"
    }
  }
}

# Read a single environment variable inline, without a data block.
output "home" {
  value = provider::environment::get_env("HOME")
}

# Provide a default to use when the variable is not set.
output "log_level" {
  value = provider::environment::get_env("LOG_LEVEL", "info")
}

# Wrap the result with sensitive() when it may contain a secret.
# NOTE: sensitive() only hides the value from plan/apply output; it is
# still written to Terraform state in plain text.
output "token" {
  value     = sensitive(provider::environment::get_env("TOKEN"))
  sensitive = true
}
