package environment

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// filterDescription documents the filter argument shared by the data source and
// the ephemeral resources.
const filterDescription = "Regular expression, in [RE2 syntax](https://github.com/google/re2/wiki/Syntax), " +
	"matched against the name of each environment variable. Only matching variables are returned. " +
	"The match is not anchored: use `^NAME$` to match a single variable. " +
	"When omitted, every environment variable is returned."

// compileFilter compiles the filter argument. A null or unknown filter returns a
// nil regular expression, which matches every variable.
func compileFilter(filter types.String) (*regexp.Regexp, diag.Diagnostics) {
	var diags diag.Diagnostics

	if filter.IsNull() || filter.IsUnknown() {
		return nil, diags
	}

	re, err := regexp.Compile(filter.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("filter"),
			"Invalid Regular Expression",
			fmt.Sprintf("The filter %q is not a valid regular expression: %s", filter.ValueString(), err),
		)
	}

	return re, diags
}

// validateFilterConfig reports an invalid filter while Terraform validates the
// configuration, so that terraform validate catches it before any plan.
func validateFilterConfig(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var filter types.String

	diags := config.GetAttribute(ctx, path.Root("filter"), &filter)

	_, filterDiags := compileFilter(filter)
	diags.Append(filterDiags...)

	return diags
}

// filterVariables converts "NAME=value" entries, as returned by os.Environ, into
// a map. When re is not nil, only the variables whose name matches are kept.
func filterVariables(environ []string, re *regexp.Regexp) map[string]string {
	items := make(map[string]string)

	for _, entry := range environ {
		// strings.Cut keeps an entry without "=" with an empty value, where the
		// SDKv2 implementation panicked on the missing value.
		name, value, _ := strings.Cut(entry, "=")

		if re != nil && !re.MatchString(name) {
			continue
		}

		items[name] = value
	}

	return items
}
