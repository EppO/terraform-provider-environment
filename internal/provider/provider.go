package environment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure environmentProvider satisfies the framework interfaces.
var (
	_ provider.Provider                       = (*environmentProvider)(nil)
	_ provider.ProviderWithEphemeralResources = (*environmentProvider)(nil)
)

// environmentProvider exposes the environment variables of the Terraform
// process. It has no configuration.
type environmentProvider struct {
	version string
}

// New returns a function that creates the provider, as expected by
// providerserver.Serve.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &environmentProvider{version: version}
	}
}

func (p *environmentProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "environment"
	resp.Version = p.version
}

func (p *environmentProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *environmentProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (p *environmentProvider) Resources(context.Context) []func() resource.Resource {
	return nil
}

func (p *environmentProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVariablesDataSource,
	}
}

func (p *environmentProvider) EphemeralResources(context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewVariablesEphemeralResource,
		NewSensitiveVariablesEphemeralResource,
	}
}
