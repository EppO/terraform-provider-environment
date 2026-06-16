package environment

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = (*getEnvMapFunction)(nil)

// getEnvMapFunction implements the provider-defined function
// provider::environment::get_env_map.
type getEnvMapFunction struct{}

// NewGetEnvMapFunction returns a new get_env_map function instance.
func NewGetEnvMapFunction() function.Function {
	return &getEnvMapFunction{}
}

func (f *getEnvMapFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "get_env_map"
}

func (f *getEnvMapFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Return environment variables as a map, optionally filtered by a regular expression.",
		MarkdownDescription: "Returns a map of environment variables. When `filter` is non-empty it is treated as a regular expression and only variables whose name matches are returned. Pass an empty string to return every variable. Wrap the result with the built-in `sensitive()` function if it may contain secrets.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "filter",
				MarkdownDescription: "Regular expression matched against variable names. An empty string returns all variables.",
			},
		},
		Return: function.MapReturn{ElementType: types.StringType},
	}
}

func (f *getEnvMapFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var filter string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &filter))
	if resp.Error != nil {
		return
	}

	items, err := filterVariables(os.Environ(), false, filter)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(
			0,
			fmt.Sprintf("invalid filter pattern %q: %s", filter, err),
		))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, items))
}
