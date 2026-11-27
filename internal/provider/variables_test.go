package environment

import (
	"maps"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCompileFilter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filter    types.String
		wantRegex string
		wantNil   bool
		wantError bool
	}{
		"null matches every variable": {
			filter:  types.StringNull(),
			wantNil: true,
		},
		"unknown is not validated": {
			filter:  types.StringUnknown(),
			wantNil: true,
		},
		"valid regular expression": {
			filter:    types.StringValue("^LC_"),
			wantRegex: "^LC_",
		},
		"invalid regular expression": {
			filter:    types.StringValue("["),
			wantError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			re, diags := compileFilter(tc.filter)

			if tc.wantError {
				if diags.ErrorsCount() != 1 {
					t.Fatalf("got diagnostics %v, want one error", diags)
				}

				d, ok := diags[0].(diag.DiagnosticWithPath)
				if !ok || !d.Path().Equal(path.Root("filter")) {
					t.Errorf("got diagnostic %v, want an error on the filter attribute", diags[0])
				}

				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if tc.wantNil != (re == nil) {
				t.Fatalf("got regular expression %v, want nil: %t", re, tc.wantNil)
			}

			if re != nil && re.String() != tc.wantRegex {
				t.Errorf("got regular expression %q, want %q", re, tc.wantRegex)
			}
		})
	}
}

func TestFilterVariables(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		environ []string
		filter  *regexp.Regexp
		want    map[string]string
	}{
		"nil filter keeps every variable": {
			environ: []string{"FOO=bar", "BAZ=qux"},
			want:    map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		"filter matches names only": {
			environ: []string{"LC_CTYPE=UTF-8", "LANG=en_US", "LC_ALL=C", "OTHER=LC_"},
			filter:  regexp.MustCompile("^LC_"),
			want:    map[string]string{"LC_CTYPE": "UTF-8", "LC_ALL": "C"},
		},
		"filter matching nothing returns an empty map": {
			environ: []string{"FOO=bar"},
			filter:  regexp.MustCompile("^NOPE$"),
			want:    map[string]string{},
		},
		"empty value is kept": {
			environ: []string{"EMPTY="},
			want:    map[string]string{"EMPTY": ""},
		},
		"value containing equal signs is kept intact": {
			environ: []string{"KEY=a=b=c"},
			want:    map[string]string{"KEY": "a=b=c"},
		},
		"entry without equal sign does not panic": {
			environ: []string{"NOEQUALS", "FOO=bar"},
			want:    map[string]string{"NOEQUALS": "", "FOO": "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := filterVariables(tc.environ, tc.filter)

			if got == nil || !maps.Equal(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}
