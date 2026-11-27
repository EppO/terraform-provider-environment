package environment

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const variablesDataSourceDeprecation = "The environment_variables data source is deprecated and will be removed in v3.0. " +
	"Terraform stores data source results in plan and state files. " +
	"Use the environment_variables or environment_sensitive_variables ephemeral resource (Terraform 1.10+) " +
	"to read variables without persisting them, or a TF_VAR_<name> input variable where ephemeral values are not allowed. " +
	"Migration guide: https://registry.terraform.io/providers/EppO/environment/latest/docs/data-sources/variables#migration"

// Ensure variablesDataSource satisfies the framework interfaces.
var (
	_ datasource.DataSource                   = (*variablesDataSource)(nil)
	_ datasource.DataSourceWithValidateConfig = (*variablesDataSource)(nil)
)

// variablesDataSource implements the deprecated environment_variables data
// source with the same schema and behavior as the SDKv2 implementation.
type variablesDataSource struct {
	environ    func() []string
	generateID func() (string, error)
}

type variablesDataSourceModel struct {
	ID        types.String      `tfsdk:"id"`
	Filter    types.String      `tfsdk:"filter"`
	Sensitive types.Bool        `tfsdk:"sensitive"`
	Items     map[string]string `tfsdk:"items"`
}

// NewVariablesDataSource returns the environment_variables data source.
func NewVariablesDataSource() datasource.DataSource {
	return &variablesDataSource{
		environ:    os.Environ,
		generateID: uuid.GenerateUUID,
	}
}

func (d *variablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variables"
}

func (d *variablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads environment variables from the environment of the Terraform process during plan. " +
			"Terraform stores the result in the plan and state. " +
			"Deprecated: use the `environment_variables` or `environment_sensitive_variables` ephemeral resource instead.",
		DeprecationMessage: variablesDataSourceDeprecation,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Random UUID, generated on every read.",
				Computed:            true,
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: filterDescription,
				Optional:            true,
			},
			"sensitive": schema.BoolAttribute{
				MarkdownDescription: "When `true`, Base64-encodes each value in `items`. " +
					"Despite its name, it does not mark the values as sensitive and they are still stored in the plan and state. " +
					"Defaults to `false`.",
				Optional: true,
			},
			"items": schema.MapAttribute{
				MarkdownDescription: "Matching environment variables, keyed by name.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *variablesDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	resp.Diagnostics.Append(validateFilterConfig(ctx, req.Config)...)
}

func (d *variablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data variablesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	re, diags := compileFilter(data.Filter)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := d.generateID()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Generate ID", err.Error())
		return
	}

	sensitive := data.Sensitive.ValueBool()
	items := filterVariables(d.environ(), re)

	if sensitive {
		for name, value := range items {
			items[name] = base64.StdEncoding.EncodeToString([]byte(value))
		}
	}

	data.ID = types.StringValue(id)
	// Store false when sensitive is not configured, like the SDKv2 default did.
	data.Sensitive = types.BoolValue(sensitive)
	data.Items = items

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
