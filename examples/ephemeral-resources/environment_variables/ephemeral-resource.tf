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
