package environment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

// protoV6ProviderFactories builds the same muxed (SDKv2 + framework) provider
// served by main.go, for use in acceptance tests across Terraform 1.8+.
var protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"environment": func() (tfprotov6.ProviderServer, error) {
		ctx := context.Background()

		upgraded, err := tf5to6server.UpgradeServer(ctx, Provider().GRPCProvider)
		if err != nil {
			return nil, err
		}

		muxServer, err := tf6muxserver.NewMuxServer(ctx,
			func() tfprotov6.ProviderServer { return upgraded },
			providerserver.NewProtocol6(NewFrameworkProvider()),
		)
		if err != nil {
			return nil, err
		}

		return muxServer.ProviderServer(), nil
	},
}
