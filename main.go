package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	environment "github.com/EppO/terraform-provider-environment/internal/provider"
)

// Generate docs for website
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

const providerAddress = "registry.terraform.io/EppO/environment"

func main() {
	ctx := context.Background()

	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// The existing data source is served by the SDKv2 provider (protocol 5);
	// upgrade it to protocol 6 so it can be muxed with the framework provider
	// that hosts the provider-defined functions.
	upgradedSDKServer, err := tf5to6server.UpgradeServer(ctx, environment.Provider().GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer { return upgradedSDKServer },
		providerserver.NewProtocol6(environment.NewFrameworkProvider()),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(providerAddress, muxServer.ProviderServer, serveOpts...); err != nil {
		log.Fatal(err)
	}
}
