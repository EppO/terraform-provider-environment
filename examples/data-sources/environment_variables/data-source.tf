# NOTE: The environment_variables data source is deprecated and will be removed in v3.0.
# Prefer the get_env / get_env_map provider-defined functions, e.g.:
#   provider::environment::get_env("HOME")
#   provider::environment::get_env_map("^LC_")

provider "environment" {}

data "environment_variables" "all" {}

data "environment_variables" "regexp" {
  filter = "^LC_"
}

data "environment_variables" "encoded" {
  filter    = "TOKEN"
  sensitive = true
}

resource "null_resource" "all" {
  triggers = data.environment_variables.all.items
}

resource "null_resource" "regexp" {
  triggers = data.environment_variables.regexp.items
}

resource "null_resource" "encoded" {
  triggers = data.environment_variables.encoded.items
}
