// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import "testing"

func TestTarget(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"valid": {
			input: "/home/user",
			want:  true,
		},

		"valid-slash": {
			input: "/home/user/",
			want:  true,
		},

		"valid-hyphen": {
			input: "/home/user-name",
			want:  true,
		},

		"max": {
			input: "/home/user1234567890",
			want:  true,
		},

		"exceed": {
			input: "/home/user1234567890123456789011234567",
			want:  false,
		},

		"special-char": {
			input: "/home/user#",
			want:  false,
		},

		"glob": {
			input: "/home/user*",
			want:  false,
		},

		"space": {
			input: "/home/ ",
			want:  false,
		},

		"space-slash": {
			input: "/home/ /",
			want:  false,
		},

		"empty": {
			input: "/home",
			want:  false,
		},

		"empty-slash": {
			input: "/home/",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := reTarget.MatchString(test.input)

			if result != test.want {
				t.Errorf("unexpected regex result %v, want %v", result, test.want)
			}
		})
	}
}
