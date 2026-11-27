package environment

import (
	"errors"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

//nolint:paralleltest // t.Setenv is not compatible with t.Parallel.
func TestAccVariablesDataSource_filter(t *testing.T) {
	t.Setenv("TF_ACC_ENV_FOO", "bar")
	t.Setenv("TF_ACC_ENV_BAZ", "qux")
	t.Setenv("TF_ACC_OTHER", "other")

	const config = `
data "environment_variables" "test" {
  filter = "^TF_ACC_ENV_"
}
`

	idChanges := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("items"), knownvalue.MapExact(map[string]knownvalue.Check{
						"TF_ACC_ENV_FOO": knownvalue.StringExact("bar"),
						"TF_ACC_ENV_BAZ": knownvalue.StringExact("qux"),
					})),
					// Like the SDKv2 default, sensitive is false when not configured.
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("sensitive"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("id"), knownvalue.StringRegexp(uuidRegexp)),
					idChanges.AddStateValue("data.environment_variables.test", tfjsonpath.New("id")),
				},
			},
			{
				// The ID is regenerated on every read.
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					idChanges.AddStateValue("data.environment_variables.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

//nolint:paralleltest // t.Setenv is not compatible with t.Parallel.
func TestAccVariablesDataSource_sensitive(t *testing.T) {
	t.Setenv("TF_ACC_ENV_FOO", "bar")
	t.Setenv("TF_ACC_ENV_BAZ", "qux")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "environment_variables" "test" {
  filter    = "^TF_ACC_ENV_"
  sensitive = true
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					// sensitive Base64-encodes the values.
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("items"), knownvalue.MapExact(map[string]knownvalue.Check{
						"TF_ACC_ENV_FOO": knownvalue.StringExact("YmFy"),
						"TF_ACC_ENV_BAZ": knownvalue.StringExact("cXV4"),
					})),
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("sensitive"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccVariablesDataSource_noMatch(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "environment_variables" "test" {
  filter = "^TF_ACC_DOES_NOT_EXIST$"
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					// An empty map, not null, like the SDKv2 implementation.
					statecheck.ExpectKnownValue("data.environment_variables.test", tfjsonpath.New("items"), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccVariablesDataSource_invalidFilter(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The SDKv2 implementation panicked on an invalid regular expression.
				Config: `
data "environment_variables" "test" {
  filter = "["
}
`,
				ExpectError: regexp.MustCompile(`Invalid Regular Expression`),
			},
		},
	})
}

// TestVariablesDataSourceRead_errors covers errors that Terraform prevents
// before calling Read or that depend on the system.
func TestVariablesDataSourceRead_errors(t *testing.T) {
	t.Parallel()

	newDataSource := func(generateID func() (string, error)) *variablesDataSource {
		return &variablesDataSource{
			environ:    func() []string { return []string{"FOO=bar"} },
			generateID: generateID,
		}
	}

	var schemaResp datasource.SchemaResponse
	newDataSource(nil).Schema(t.Context(), datasource.SchemaRequest{}, &schemaResp)
	objectType := schemaResp.Schema.Type().TerraformType(t.Context())

	readConfig := func(filter string) tftypes.Value {
		return tftypes.NewValue(objectType, map[string]tftypes.Value{
			"id":        tftypes.NewValue(tftypes.String, nil),
			"filter":    tftypes.NewValue(tftypes.String, filter),
			"sensitive": tftypes.NewValue(tftypes.Bool, nil),
			"items":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		})
	}

	tests := map[string]struct {
		dataSource *variablesDataSource
		config     tftypes.Value
		want       string
	}{
		"configuration not matching the schema": {
			dataSource: newDataSource(nil),
			config:     tftypes.NewValue(tftypes.String, "invalid"),
			want:       "Value Conversion Error",
		},
		"invalid filter": {
			dataSource: newDataSource(nil),
			config:     readConfig("["),
			want:       "Invalid Regular Expression",
		},
		"ID generation failure": {
			dataSource: newDataSource(func() (string, error) { return "", errors.New("no entropy") }),
			config:     readConfig("^FOO$"),
			want:       "Unable to Generate ID",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := datasource.ReadRequest{
				Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tc.config},
			}
			resp := datasource.ReadResponse{
				State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, nil)},
			}

			tc.dataSource.Read(t.Context(), req, &resp)

			if !hasErrorSummary(resp.Diagnostics, tc.want) {
				t.Errorf("got diagnostics %v, want an error %q", resp.Diagnostics, tc.want)
			}

			if !resp.State.Raw.IsNull() {
				t.Errorf("got state %s, want null", resp.State.Raw)
			}
		})
	}
}
