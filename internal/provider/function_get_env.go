package environment

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = (*getEnvFunction)(nil)

// getEnvFunction implements the provider-defined function
// provider::environment::get_env.
type getEnvFunction struct{}

// NewGetEnvFunction returns a new get_env function instance.
func NewGetEnvFunction() function.Function {
	return &getEnvFunction{}
}

func (f *getEnvFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "get_env"
}

func (f *getEnvFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Return the value of a single environment variable.",
		MarkdownDescription: "Returns the value of the environment variable `name`. If the variable is not set, the first `default` argument is returned instead; if no default is provided, the function returns an error.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "name",
				MarkdownDescription: "Name of the environment variable to read.",
			},
		},
		VariadicParameter: function.StringParameter{
			Name:                "default",
			MarkdownDescription: "Optional value to return when the environment variable is not set. Only the first value is used.",
		},
		Return: function.StringReturn{},
	}
}

func (f *getEnvFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name string
	var defaults []string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &name, &defaults))
	if resp.Error != nil {
		return
	}

	value, ok := os.LookupEnv(name)
	if !ok {
		if len(defaults) == 0 {
			resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(
				0,
				fmt.Sprintf("environment variable %q is not set and no default value was provided", name),
			))
			return
		}
		value = defaults[0]
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, value))
}
