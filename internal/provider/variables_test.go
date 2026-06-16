package environment

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestFilterVariables(t *testing.T) {
	t.Parallel()

	tokenB64 := base64.StdEncoding.EncodeToString([]byte("secret"))

	tests := map[string]struct {
		variables []string
		sensitive bool
		filter    string
		want      map[string]string
		wantErr   bool
	}{
		"no filter returns all": {
			variables: []string{"FOO=bar", "BAZ=qux"},
			want:      map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		"regex filter keeps matches only": {
			variables: []string{"LC_CTYPE=UTF-8", "LANG=en", "LC_ALL=C"},
			filter:    "^LC_",
			want:      map[string]string{"LC_CTYPE": "UTF-8", "LC_ALL": "C"},
		},
		"filter matching nothing returns empty map": {
			variables: []string{"FOO=bar"},
			filter:    "^NOPE$",
			want:      map[string]string{},
		},
		"sensitive base64-encodes values": {
			variables: []string{"TOKEN=secret"},
			sensitive: true,
			filter:    "TOKEN",
			want:      map[string]string{"TOKEN": tokenB64},
		},
		"empty value is preserved": {
			variables: []string{"EMPTY="},
			want:      map[string]string{"EMPTY": ""},
		},
		"value containing equals is kept intact": {
			variables: []string{"KEY=a=b=c"},
			want:      map[string]string{"KEY": "a=b=c"},
		},
		"entry without equals does not panic": {
			variables: []string{"NOEQUALS", "FOO=bar"},
			want:      map[string]string{"NOEQUALS": "", "FOO": "bar"},
		},
		"invalid regex returns error": {
			variables: []string{"FOO=bar"},
			filter:    "[",
			wantErr:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := filterVariables(tc.variables, tc.sensitive, tc.filter)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}
