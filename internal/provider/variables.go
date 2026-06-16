package environment

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// filterVariables turns a slice of "NAME=value" strings (as produced by
// os.Environ()) into a map. When filter is non-empty it is compiled as a regular
// expression and only variables whose name matches are kept; an invalid pattern
// returns an error. When sensitive is true each value is base64-encoded.
//
// This is the shared core used by both the environment_variables data source and
// the provider-defined functions so that filtering and encoding stay identical.
func filterVariables(variables []string, sensitive bool, filter string) (map[string]string, error) {
	var re *regexp.Regexp
	if filter != "" {
		var err error
		re, err = regexp.Compile(filter)
		if err != nil {
			return nil, err
		}
	}

	out := make(map[string]string)
	for _, variable := range variables {
		// strings.Cut guards against entries with no "=" (found == false),
		// which would otherwise panic on a SplitN index access.
		name, value, _ := strings.Cut(variable, "=")

		if re != nil && !re.MatchString(name) {
			continue
		}
		if sensitive {
			value = base64.StdEncoding.EncodeToString([]byte(value))
		}

		out[name] = value
	}

	return out, nil
}
