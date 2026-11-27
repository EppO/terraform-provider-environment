package environment

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories serves the provider for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"environment": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho adds the echo provider, which copies
// its configuration into the state of an echo resource. Ephemeral values can
// configure a provider, so this is how acceptance tests observe them.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"environment": providerserver.NewProtocol6WithError(New("test")()),
	"echo":        echoprovider.NewProviderServer(),
}

// newProviderServer returns a provider server and the schemas it reports.
func newProviderServer(t *testing.T) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()

	server, err := testAccProtoV6ProviderFactories["environment"]()
	if err != nil {
		t.Fatalf("unable to create provider server: %s", err)
	}

	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("unable to get provider schema: %s", err)
	}

	if len(schemas.Diagnostics) > 0 {
		t.Fatalf("unable to get provider schema: %v", schemas.Diagnostics)
	}

	return server, schemas
}

// TestProviderSchema checks which items are tagged as sensitive in the schemas
// reported to Terraform.
func TestProviderSchema(t *testing.T) {
	t.Parallel()

	_, schemas := newProviderServer(t)

	tests := map[string]struct {
		schema        *tfprotov6.Schema
		wantSensitive bool
	}{
		// Marking the data source items as sensitive would break configurations
		// that output them without sensitive = true.
		"data source": {
			schema: schemas.DataSourceSchemas["environment_variables"],
		},
		"ephemeral resource": {
			schema: schemas.EphemeralResourceSchemas["environment_variables"],
		},
		"sensitive ephemeral resource": {
			schema:        schemas.EphemeralResourceSchemas["environment_sensitive_variables"],
			wantSensitive: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.schema == nil {
				t.Fatal("missing schema")
			}

			i := slices.IndexFunc(tc.schema.Block.Attributes, func(a *tfprotov6.SchemaAttribute) bool { return a.Name == "items" })
			if i < 0 {
				t.Fatal("missing items attribute")
			}

			if got := tc.schema.Block.Attributes[i].Sensitive; got != tc.wantSensitive {
				t.Errorf("got items sensitive %t, want %t", got, tc.wantSensitive)
			}
		})
	}
}

// TestProviderValidateConfig checks the diagnostics returned when Terraform
// validates a configuration, as terraform validate does.
func TestProviderValidateConfig(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	server, schemas := newProviderServer(t)

	validateDataSource := func(typeName string, config *tfprotov6.DynamicValue) ([]*tfprotov6.Diagnostic, error) {
		resp, err := server.ValidateDataResourceConfig(ctx, &tfprotov6.ValidateDataResourceConfigRequest{TypeName: typeName, Config: config})
		if err != nil {
			return nil, err
		}

		return resp.Diagnostics, nil
	}

	validateEphemeralResource := func(typeName string, config *tfprotov6.DynamicValue) ([]*tfprotov6.Diagnostic, error) {
		resp, err := server.ValidateEphemeralResourceConfig(ctx, &tfprotov6.ValidateEphemeralResourceConfigRequest{TypeName: typeName, Config: config})
		if err != nil {
			return nil, err
		}

		return resp.Diagnostics, nil
	}

	tests := map[string]struct {
		schema   *tfprotov6.Schema
		validate func(string, *tfprotov6.DynamicValue) ([]*tfprotov6.Diagnostic, error)
		typeName string
		filter   tftypes.Value
		want     []string
	}{
		"data source warns about its deprecation": {
			schema:   schemas.DataSourceSchemas["environment_variables"],
			validate: validateDataSource,
			typeName: "environment_variables",
			filter:   tftypes.NewValue(tftypes.String, "^HOME$"),
			want:     []string{"WARNING: Deprecated"},
		},
		"data source rejects an invalid filter": {
			schema:   schemas.DataSourceSchemas["environment_variables"],
			validate: validateDataSource,
			typeName: "environment_variables",
			filter:   tftypes.NewValue(tftypes.String, "["),
			want:     []string{"ERROR: Invalid Regular Expression", "WARNING: Deprecated"},
		},
		"ephemeral resource accepts a valid filter": {
			schema:   schemas.EphemeralResourceSchemas["environment_variables"],
			validate: validateEphemeralResource,
			typeName: "environment_variables",
			filter:   tftypes.NewValue(tftypes.String, "^HOME$"),
		},
		"ephemeral resource accepts an unknown filter": {
			schema:   schemas.EphemeralResourceSchemas["environment_variables"],
			validate: validateEphemeralResource,
			typeName: "environment_variables",
			filter:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"ephemeral resource rejects an invalid filter": {
			schema:   schemas.EphemeralResourceSchemas["environment_variables"],
			validate: validateEphemeralResource,
			typeName: "environment_variables",
			filter:   tftypes.NewValue(tftypes.String, "["),
			want:     []string{"ERROR: Invalid Regular Expression"},
		},
		"sensitive ephemeral resource rejects an invalid filter": {
			schema:   schemas.EphemeralResourceSchemas["environment_sensitive_variables"],
			validate: validateEphemeralResource,
			typeName: "environment_sensitive_variables",
			filter:   tftypes.NewValue(tftypes.String, "["),
			want:     []string{"ERROR: Invalid Regular Expression"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags, err := tc.validate(tc.typeName, configWithFilter(t, tc.schema, tc.filter))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			var got []string
			for _, d := range diags {
				got = append(got, d.Severity.String()+": "+d.Summary)
			}
			slices.Sort(got)

			if !slices.Equal(got, tc.want) {
				t.Errorf("got diagnostics %q, want %q", got, tc.want)
			}
		})
	}
}

// configWithFilter returns a configuration for the given schema where every
// attribute is null except filter.
func configWithFilter(t *testing.T, schema *tfprotov6.Schema, filter tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()

	if schema == nil {
		t.Fatal("missing schema")
	}

	objectType, ok := schema.ValueType().(tftypes.Object)
	if !ok {
		t.Fatalf("schema type %s is not an object", schema.ValueType())
	}

	values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}
	values["filter"] = filter

	config, err := tfprotov6.NewDynamicValue(objectType, tftypes.NewValue(objectType, values))
	if err != nil {
		t.Fatalf("unable to create configuration: %s", err)
	}

	return &config
}

// hasErrorSummary reports whether diags contain an error with the given summary.
func hasErrorSummary(diags diag.Diagnostics, summary string) bool {
	for _, d := range diags.Errors() {
		if d.Summary() == summary {
			return true
		}
	}

	return false
}
