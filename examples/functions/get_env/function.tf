# Read a single environment variable inline, without a data block.
output "home" {
  value = provider::environment::get_env("HOME")
}

# Provide a default to use when the variable is not set.
output "log_level" {
  value = provider::environment::get_env("LOG_LEVEL", "info")
}
