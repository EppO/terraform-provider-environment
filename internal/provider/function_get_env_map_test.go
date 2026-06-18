package environment

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runGetEnvMap(t *testing.T, filter string) *function.RunResponse {
	t.Helper()

	req := function.RunRequest{
		Arguments: function.NewArgumentsData([]attr.Value{types.StringValue(filter)}),
	}
	resp := &function.RunResponse{
		Result: function.NewResultData(types.MapNull(types.StringType)),
	}

	NewGetEnvMapFunction().Run(context.Background(), req, resp)
	return resp
}

//nolint:paralleltest // uses t.Setenv
func TestGetEnvMapFunction_FilterReturnsMatches(t *testing.T) {
	t.Setenv("ENV_MAP_TEST_ONE", "1")
	t.Setenv("ENV_MAP_TEST_TWO", "2")

	resp := runGetEnvMap(t, "^ENV_MAP_TEST_")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error)
	}

	elements := resp.Result.Value().(types.Map).Elements()
	if len(elements) != 2 {
		t.Fatalf("got %d elements, want 2: %#v", len(elements), elements)
	}
	if got := elements["ENV_MAP_TEST_ONE"].(types.String).ValueString(); got != "1" {
		t.Errorf("ENV_MAP_TEST_ONE = %q, want %q", got, "1")
	}
}

//nolint:paralleltest // depends on process environment
func TestGetEnvMapFunction_NonMatchingFilterReturnsEmpty(t *testing.T) {
	resp := runGetEnvMap(t, "^ENV_MAP_TEST_NOPE_")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if n := len(resp.Result.Value().(types.Map).Elements()); n != 0 {
		t.Errorf("got %d elements, want 0", n)
	}
}

//nolint:paralleltest // depends on process environment
func TestGetEnvMapFunction_InvalidRegexReturnsError(t *testing.T) {
	resp := runGetEnvMap(t, "[")
	if resp.Error == nil {
		t.Fatalf("expected error for invalid regex, got nil")
	}
}
