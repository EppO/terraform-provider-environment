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
