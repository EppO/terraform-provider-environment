package environment

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Ephemeral resources require Terraform 1.10 or later.
var ephemeralResourceVersionChecks = []tfversion.TerraformVersionCheck{
	tfversion.SkipBelow(tfversion.Version1_10_0),
}

// ephemeralResourceTypes are the ephemeral resource types. They only differ by
// the sensitivity of their items in the schema, see TestProviderSchema.
var ephemeralResourceTypes = []string{"environment_variables", "environment_sensitive_variables"}

//nolint:paralleltest // t.Setenv is not compatible with t.Parallel.
func TestAccVariablesEphemeralResource_filter(t *testing.T) {
	for _, typeName := range ephemeralResourceTypes {
		t.Run(typeName, func(t *testing.T) {
			t.Setenv("TF_ACC_ENV_FOO", "bar")
			t.Setenv("TF_ACC_ENV_BAZ", "qux")
			t.Setenv("TF_ACC_OTHER", "other")

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				TerraformVersionChecks:   ephemeralResourceVersionChecks,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
ephemeral %[1]q "test" {
  filter = "^TF_ACC_ENV_"
}

provider "echo" {
  data = {
    items = ephemeral.%[1]s.test.items
    # ephemeralasnull only returns null for ephemeral values.
    ephemeral = ephemeralasnull(ephemeral.%[1]s.test.items) == null
  }
}

resource "echo" "test" {}
`, typeName),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("echo.test", tfjsonpath.New("data"), knownvalue.ObjectExact(map[string]knownvalue.Check{
								"items": knownvalue.MapExact(map[string]knownvalue.Check{
									"TF_ACC_ENV_FOO": knownvalue.StringExact("bar"),
									"TF_ACC_ENV_BAZ": knownvalue.StringExact("qux"),
								}),
								"ephemeral": knownvalue.Bool(true),
							})),
						},
					},
				},
			})
		})
	}
}

//nolint:paralleltest // t.Setenv is not compatible with t.Parallel.
func TestAccVariablesEphemeralResource_noFilter(t *testing.T) {
	t.Setenv("TF_ACC_ENV_FOO", "bar")
	t.Setenv("TF_ACC_OTHER", "other")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		TerraformVersionChecks:   ephemeralResourceVersionChecks,
		Steps: []resource.TestStep{
			{
				// Only two variables are copied, to keep the rest of the
				// environment out of the test state.
				Config: `
ephemeral "environment_variables" "test" {}

provider "echo" {
  data = {
    foo   = ephemeral.environment_variables.test.items["TF_ACC_ENV_FOO"]
    other = ephemeral.environment_variables.test.items["TF_ACC_OTHER"]
  }
}

resource "echo" "test" {}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test", tfjsonpath.New("data"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"foo":   knownvalue.StringExact("bar"),
						"other": knownvalue.StringExact("other"),
					})),
				},
			},
		},
	})
}

func TestAccVariablesEphemeralResource_noMatch(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		TerraformVersionChecks:   ephemeralResourceVersionChecks,
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "environment_variables" "test" {
  filter = "^TF_ACC_DOES_NOT_EXIST$"
}

provider "echo" {
  data = ephemeral.environment_variables.test.items
}

resource "echo" "test" {}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test", tfjsonpath.New("data"), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccVariablesEphemeralResource_invalidFilter(t *testing.T) {
	t.Parallel()

	for _, typeName := range ephemeralResourceTypes {
		t.Run(typeName, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				TerraformVersionChecks:   ephemeralResourceVersionChecks,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
ephemeral %q "test" {
  filter = "["
}
`, typeName),
						ExpectError: regexp.MustCompile(`Invalid Regular Expression`),
					},
				},
			})
		})
	}
}

// TestAccVariablesEphemeralResource_readAtApply changes a variable after the
// plan is saved: Terraform must read it again when applying the plan instead of
// reusing the planned value, which is what the data source does.
//
//nolint:paralleltest // t.Setenv is not compatible with t.Parallel.
func TestAccVariablesEphemeralResource_readAtApply(t *testing.T) {
	for _, typeName := range ephemeralResourceTypes {
		t.Run(typeName, func(t *testing.T) {
			t.Setenv("TF_ACC_ENV_PHASE", "plan")

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				TerraformVersionChecks:   ephemeralResourceVersionChecks,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
ephemeral %[1]q "test" {
  filter = "^TF_ACC_ENV_PHASE$"
}

provider "echo" {
  data = ephemeral.%[1]s.test.items["TF_ACC_ENV_PHASE"]
}

resource "echo" "test" {}
`, typeName),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								setenvCheck{t: t, key: "TF_ACC_ENV_PHASE", value: "apply"},
							},
						},
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("echo.test", tfjsonpath.New("data"), knownvalue.StringExact("apply")),
						},
					},
				},
			})
		})
	}
}

// setenvCheck is a plan check that sets an environment variable, to change the
// environment between the plan and the apply of a test step.
type setenvCheck struct {
	t     *testing.T
	key   string
	value string
}

func (c setenvCheck) CheckPlan(context.Context, plancheck.CheckPlanRequest, *plancheck.CheckPlanResponse) {
	c.t.Setenv(c.key, c.value)
}

// TestVariablesEphemeralResourceOpen_errors covers errors that Terraform
// prevents before calling Open.
func TestVariablesEphemeralResourceOpen_errors(t *testing.T) {
	t.Parallel()

	r := &variablesEphemeralResource{environ: func() []string { return []string{"FOO=bar"} }}

	var schemaResp ephemeral.SchemaResponse
	r.Schema(t.Context(), ephemeral.SchemaRequest{}, &schemaResp)
	objectType := schemaResp.Schema.Type().TerraformType(t.Context())

	tests := map[string]struct {
		config tftypes.Value
		want   string
	}{
		"configuration not matching the schema": {
			config: tftypes.NewValue(tftypes.String, "invalid"),
			want:   "Value Conversion Error",
		},
		"invalid filter": {
			config: tftypes.NewValue(objectType, map[string]tftypes.Value{
				"filter": tftypes.NewValue(tftypes.String, "["),
				"items":  tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			}),
			want: "Invalid Regular Expression",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := ephemeral.OpenRequest{
				Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tc.config},
			}
			resp := ephemeral.OpenResponse{
				Result: tfsdk.EphemeralResultData{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, nil)},
			}

			r.Open(t.Context(), req, &resp)

			if !hasErrorSummary(resp.Diagnostics, tc.want) {
				t.Errorf("got diagnostics %v, want an error %q", resp.Diagnostics, tc.want)
			}

			if !resp.Result.Raw.IsNull() {
				t.Errorf("got result %s, want null", resp.Result.Raw)
			}
		})
	}
}
