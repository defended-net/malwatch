// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import (
	"testing"
)

func TestReMetaChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid",
			input: "malwatch123",
			want:  false,
		},
		{
			name:  "semicolon",
			input: "mal;watch",
			want:  true,
		},
		{
			name:  "amp",
			input: "mal&watch",
			want:  true,
		},
		{
			name:  "pipe",
			input: "mal|watch",
			want:  true,
		},
		{
			name:  "dollar",
			input: "mal$watch",
			want:  true,
		},
		{
			name:  "backslash",
			input: "mal\\watch",
			want:  true,
		},
		{
			name:  "bracket-open",
			input: "mal(watch",
			want:  true,
		},
		{
			name:  "bracket-close",
			input: "mal)watch",
			want:  true,
		},
		{
			name:  "smaller",
			input: "mal<watch",
			want:  true,
		},
		{
			name:  "greater",
			input: "mal>watch",
			want:  true,
		},
		{
			name:  "empty",
			input: "",
			want:  false,
		},
		{
			name:  "tab",
			input: "mal\twatch",
			want:  false,
		},
		{
			name:  "mixed",
			input: "abc!@#%^*_+-=[]{}':\",./?",
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := reMetaChars.MatchString(test.input)

			if got != test.want {
				t.Errorf("unexpected metachars result %v, want %v", got, test.want)
			}
		})
	}
}
