---
layout: ""
page_title: "Provider: Environment"
description: |-
  The Environment provider reads the environment variables of the Terraform process through ephemeral resources, which Terraform never stores in plan or state files.
---

# Environment Provider

The Environment provider reads the environment variables of the Terraform process,
such as the variables of the CI/CD job running Terraform.

Use the [`environment_variables`](./ephemeral-resources/variables.md) and
[`environment_sensitive_variables`](./ephemeral-resources/sensitive_variables.md)
ephemeral resources to read them. Terraform reads ephemeral resources again in every
plan and apply, and never stores their values in plan or state files.

~> **Deprecated:** The `environment_variables` data source is deprecated and will be
removed in **v3.0**, because Terraform stores data source results in plan and state
files. See the [data source documentation](./data-sources/variables.md#migration)
for a migration guide.

## Requirements

| Feature | Terraform version |
| --- | --- |
| Provider and `environment_variables` data source (deprecated) | 1.0 or later |
| `environment_variables` and `environment_sensitive_variables` ephemeral resources | 1.10 or later |
| Write-only arguments, to pass ephemeral values to managed resources | 1.11 or later |

## Example Usage

```terraform
terraform {
  # Ephemeral resources require Terraform 1.10 or later.
  required_version = ">= 1.10"
}

# Reads the deployment settings exported by the CI/CD job running Terraform.
ephemeral "environment_variables" "deploy" {
  filter = "^DEPLOY_"
}

provider "aws" {
  # lookup returns a default value when the variable isn't set.
  region = lookup(ephemeral.environment_variables.deploy.items, "DEPLOY_REGION", "us-east-1")

  assume_role {
    role_arn = ephemeral.environment_variables.deploy.items["DEPLOY_ROLE_ARN"]
  }
}
```

Read secrets with `environment_sensitive_variables`:

```terraform
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
```

## Schema

The provider has no configuration arguments.
