package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	environment "github.com/EppO/terraform-provider-environment/internal/provider"
)

// Generate docs for website
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

const providerAddress = "registry.terraform.io/EppO/environment"

// version is set by goreleaser through -ldflags "-X main.version=...".
var version = "dev"

// serve and fatal are variables so that tests can run main without starting a
// gRPC server or exiting the test binary.
var (
	serve = providerserver.Serve
	fatal = log.Fatal
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: providerAddress,
		Debug:   debug,
	}

	if err := serve(context.Background(), environment.New(version), opts); err != nil {
		fatal(err)
	}
}
