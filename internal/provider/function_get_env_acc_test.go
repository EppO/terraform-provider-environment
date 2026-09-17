package environment

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

//nolint:paralleltest // uses t.Setenv
func TestAccGetEnvFunction(t *testing.T) {
	t.Setenv("ENV_ACC_GET_ENV", "value-from-env")

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `output "test" { value = provider::environment::get_env("ENV_ACC_GET_ENV") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("value-from-env")),
				},
			},
			{
				Config: `output "test" { value = provider::environment::get_env("ENV_ACC_DEFINITELY_UNSET", "fallback") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("fallback")),
				},
			},
		},
	})
}

//nolint:paralleltest // uses t.Setenv
func TestAccGetEnvMapFunction(t *testing.T) {
	t.Setenv("ENV_ACC_MAP_ONE", "1")
	t.Setenv("ENV_ACC_MAP_TWO", "2")

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `output "test" { value = provider::environment::get_env_map("^ENV_ACC_MAP_") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.MapExact(map[string]knownvalue.Check{
						"ENV_ACC_MAP_ONE": knownvalue.StringExact("1"),
						"ENV_ACC_MAP_TWO": knownvalue.StringExact("2"),
					})),
				},
			},
		},
	})
}

//nolint:paralleltest // uses t.Setenv
func TestAccVariablesDataSource(t *testing.T) {
	t.Setenv("ENV_ACC_DS_VAR", "ds-value")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "environment_variables" "test" { filter = "^ENV_ACC_DS_VAR$" }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.environment_variables.test",
						tfjsonpath.New("items").AtMapKey("ENV_ACC_DS_VAR"),
						knownvalue.StringExact("ds-value"),
					),
				},
			},
		},
	})
}
