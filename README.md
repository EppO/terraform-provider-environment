# Terraform Provider Environment [![release](https://github.com/EppO/terraform-provider-environment/actions/workflows/release.yml/badge.svg)](https://github.com/EppO/terraform-provider-environment/actions/workflows/release.yml)

Terraform provider that reads the environment variables of the Terraform process,
such as the variables of the CI/CD job running Terraform.

It provides two ephemeral resources (Terraform 1.10 or later). Terraform reads them
again in every plan and apply, and never stores their values in plan or state files:

- `environment_variables` reads environment variables, optionally filtered by a
  regular expression.
- `environment_sensitive_variables` reads secrets: the provider schema marks its
  values as sensitive.

The `environment_variables` data source is deprecated and will be removed in v3.0,
because Terraform stores its results in plan and state files. See the
[migration guide](docs/data-sources/variables.md#migration).

## Requirements

| Feature | Terraform version |
| --- | --- |
| Provider and `environment_variables` data source (deprecated) | 1.0 or later |
| `environment_variables` and `environment_sensitive_variables` ephemeral resources | 1.10 or later |
| Write-only arguments, to pass ephemeral values to managed resources | 1.11 or later |

## Example

```hcl
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

More examples are available in the [examples](examples) directory and in the
[documentation](docs).

## Test

```shell
make test     # unit tests
make testacc  # unit and acceptance tests, requires Terraform
make cover    # unit and acceptance tests, with a coverage report
```

The acceptance tests of the ephemeral resources are skipped with Terraform versions
older than 1.10.

## Build

Run the following command to build the provider

```shell
make build
```

## Install

```shell
make install
```
