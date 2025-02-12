// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"strings"
	"testing"
)

func TestYesNo(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"empty": {
			input: "",
			want:  false,
		},

		"space": {
			input: " ",
			want:  false,
		},

		"new-line": {
			input: "\n",
			want:  false,
		},

		"Y": {
			input: "Y",
			want:  true,
		},

		"y": {
			input: "y",
			want:  true,
		},

		"N": {
			input: "N",
			want:  false,
		},

		"n": {
			input: "n",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(test.input)

			got := YesNo(t.Name(), reader)

			if got != test.want {
				t.Fatalf("unexpected yesno val want %v", test.want)
			}
		})
	}
}
