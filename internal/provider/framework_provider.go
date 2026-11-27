package environment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure frameworkProvider satisfies the framework provider interfaces.
var (
	_ provider.Provider              = (*frameworkProvider)(nil)
	_ provider.ProviderWithFunctions = (*frameworkProvider)(nil)
)

// frameworkProvider is a terraform-plugin-framework provider whose sole purpose
// is to expose provider-defined functions (Terraform 1.8+). The existing data
// source continues to be served by the SDKv2 provider; the two are combined with
// terraform-plugin-mux in main.go.
type frameworkProvider struct{}

// NewFrameworkProvider returns the framework provider used for muxing.
func NewFrameworkProvider() provider.Provider {
	return &frameworkProvider{}
}

func (p *frameworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "environment"
}

func (p *frameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *frameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *frameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *frameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *frameworkProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewGetEnvFunction,
		NewGetEnvMapFunction,
	}
}
