terraform {
  # Write-only arguments require Terraform 1.11 or later.
  required_version = ">= 1.11"
}

# Reads the database password from the DB_PASSWORD environment variable.
ephemeral "environment_sensitive_variables" "database" {
  filter = "^DB_PASSWORD$"
}

resource "aws_db_instance" "example" {
  identifier          = "example"
  engine              = "postgres"
  instance_class      = "db.t4g.micro"
  allocated_storage   = 20
  username            = "app"
  skip_final_snapshot = true

  # Write-only arguments accept ephemeral values, and Terraform never stores them.
  password_wo = ephemeral.environment_sensitive_variables.database.items["DB_PASSWORD"]
  # Terraform can't detect changes to write-only arguments: increment the
  # version to update the password.
  password_wo_version = 1
}
