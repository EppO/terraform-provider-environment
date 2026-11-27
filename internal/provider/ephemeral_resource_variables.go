package environment

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure variablesEphemeralResource satisfies the framework interfaces.
var (
	_ ephemeral.EphemeralResource                   = (*variablesEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithValidateConfig = (*variablesEphemeralResource)(nil)
)

// variablesEphemeralResource implements the environment_variables and the
// environment_sensitive_variables ephemeral resources. Sensitivity is a static
// property of a schema, so secrets get a dedicated resource type whose items
// attribute is marked as sensitive.
type variablesEphemeralResource struct {
	sensitive bool
	environ   func() []string
}

type variablesEphemeralResourceModel struct {
	Filter types.String      `tfsdk:"filter"`
	Items  map[string]string `tfsdk:"items"`
}

// NewVariablesEphemeralResource returns the environment_variables ephemeral
// resource.
func NewVariablesEphemeralResource() ephemeral.EphemeralResource {
	return &variablesEphemeralResource{environ: os.Environ}
}

// NewSensitiveVariablesEphemeralResource returns the
// environment_sensitive_variables ephemeral resource.
func NewSensitiveVariablesEphemeralResource() ephemeral.EphemeralResource {
	return &variablesEphemeralResource{sensitive: true, environ: os.Environ}
}

func (r *variablesEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	if r.sensitive {
		resp.TypeName = req.ProviderTypeName + "_sensitive_variables"
		return
	}

	resp.TypeName = req.ProviderTypeName + "_variables"
}

func (r *variablesEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	description := "Reads environment variables from the environment of the Terraform process. " +
		"Terraform reads them again in every plan and apply, and never stores them in plan or state files. " +
		"Requires Terraform 1.10 or later. " +
		"To read secrets, use `environment_sensitive_variables`, which marks the values as sensitive in the provider schema."

	if r.sensitive {
		description = "Reads secrets, such as tokens or passwords, from the environment of the Terraform process. " +
			"Same as `environment_variables`, except that the provider schema marks `items` as sensitive. " +
			"Terraform reads them again in every plan and apply, and never stores them in plan or state files. " +
			"Requires Terraform 1.10 or later."
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{
				MarkdownDescription: filterDescription,
				Optional:            true,
			},
			"items": schema.MapAttribute{
				MarkdownDescription: "Matching environment variables, keyed by name.",
				ElementType:         types.StringType,
				Computed:            true,
				Sensitive:           r.sensitive,
			},
		},
	}
}

func (r *variablesEphemeralResource) ValidateConfig(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	resp.Diagnostics.Append(validateFilterConfig(ctx, req.Config)...)
}

func (r *variablesEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data variablesEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	re, diags := compileFilter(data.Filter)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Items = filterVariables(r.environ(), re)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
