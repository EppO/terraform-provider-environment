package environment

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// runGetEnv invokes the get_env function with the given name argument and
// optional default values, returning the response.
func runGetEnv(t *testing.T, name string, defaults ...string) *function.RunResponse {
	t.Helper()

	defaultTypes := make([]attr.Type, len(defaults))
	defaultValues := make([]attr.Value, len(defaults))
	for i, d := range defaults {
		defaultTypes[i] = types.StringType
		defaultValues[i] = types.StringValue(d)
	}

	req := function.RunRequest{
		Arguments: function.NewArgumentsData([]attr.Value{
			types.StringValue(name),
			types.TupleValueMust(defaultTypes, defaultValues),
		}),
	}
	resp := &function.RunResponse{Result: function.NewResultData(types.StringNull())}

	NewGetEnvFunction().Run(context.Background(), req, resp)
	return resp
}

// These tests use t.Setenv and therefore cannot run in parallel.

//nolint:paralleltest // uses t.Setenv
func TestGetEnvFunction_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("ENV_FUNC_TEST_SET", "hello")

	resp := runGetEnv(t, "ENV_FUNC_TEST_SET")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if got := resp.Result.Value().(types.String).ValueString(); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

//nolint:paralleltest // uses t.Setenv
func TestGetEnvFunction_ReturnsEmptyStringForEmptyValue(t *testing.T) {
	t.Setenv("ENV_FUNC_TEST_EMPTY", "")

	resp := runGetEnv(t, "ENV_FUNC_TEST_EMPTY", "fallback")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if got := resp.Result.Value().(types.String).ValueString(); got != "" {
		t.Errorf("got %q, want empty string (set-but-empty must not use default)", got)
	}
}

//nolint:paralleltest // depends on process environment
func TestGetEnvFunction_ReturnsDefaultWhenUnset(t *testing.T) {
	resp := runGetEnv(t, "ENV_FUNC_TEST_DEFINITELY_UNSET", "fallback")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if got := resp.Result.Value().(types.String).ValueString(); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

//nolint:paralleltest // depends on process environment
func TestGetEnvFunction_ErrorsWhenUnsetWithoutDefault(t *testing.T) {
	resp := runGetEnv(t, "ENV_FUNC_TEST_DEFINITELY_UNSET")
	if resp.Error == nil {
		t.Fatalf("expected error for unset variable with no default, got nil")
	}
}
