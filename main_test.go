package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//nolint:paralleltest // main uses os.Args, the global flag set and package-level hooks.
func TestMainServesProvider(t *testing.T) {
	originalArgs, originalFlags, originalServe, originalFatal := os.Args, flag.CommandLine, serve, fatal
	t.Cleanup(func() {
		os.Args, flag.CommandLine, serve, fatal = originalArgs, originalFlags, originalServe, originalFatal
	})

	os.Args = []string{"terraform-provider-environment", "-debug"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	serveErr := errors.New("serve failed")

	var (
		gotProvider provider.Provider
		gotOpts     providerserver.ServeOpts
		gotFatal    []any
	)

	serve = func(_ context.Context, providerFunc func() provider.Provider, opts providerserver.ServeOpts) error {
		gotProvider, gotOpts = providerFunc(), opts
		return serveErr
	}
	fatal = func(v ...any) { gotFatal = v }

	main()

	if want := "registry.terraform.io/EppO/environment"; gotOpts.Address != want {
		t.Errorf("got address %q, want %q", gotOpts.Address, want)
	}

	if !gotOpts.Debug {
		t.Error("got debug false, want true with -debug")
	}

	var metadata provider.MetadataResponse
	gotProvider.Metadata(t.Context(), provider.MetadataRequest{}, &metadata)

	if metadata.TypeName != "environment" || metadata.Version != version {
		t.Errorf("got provider %q version %q, want %q version %q", metadata.TypeName, metadata.Version, "environment", version)
	}

	if len(gotFatal) != 1 || gotFatal[0] != serveErr {
		t.Errorf("got fatal %v, want %v", gotFatal, serveErr)
	}
}
