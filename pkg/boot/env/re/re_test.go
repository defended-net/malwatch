// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package re

import (
	"regexp"
	"slices"
	"testing"
)

func TestMain(m *testing.M) {
	targets = []*regexp.Regexp{
		regexp.MustCompile(`^/(?P<target>target)/?(.*)`),
	}

	m.Run()
}

func TestRuleName(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"word": {
			input: `rule`,
			want:  true,
		},

		"word-num": {
			input: `rule0`,
			want:  true,
		},

		"word-underscore": {
			input: `rule_`,
			want:  true,
		},

		"word-underscore-num": {
			input: `rule_0`,
			want:  true,
		},

		"star": {
			input: `*`,
			want:  true,
		},

		"word-hyphen": {
			input: `rule-`,
			want:  false,
		},

		"word-space": {
			input: `rule `,
			want:  false,
		},

		"empty": {
			input: ``,
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := yrName.MatchString(test.input)

			if result != test.want {
				t.Errorf("unexpected regex result %v, want %v", result, test.want)
			}
		})
	}
}

func TestTargets(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"valid-file": {
			input: "/target/index.php",
			want:  "target",
		},

		"invalid-file": {
			input: "/no-target/index.php",
			want:  "fs",
		},

		"valid-dir": {
			input: "/target/dir",
			want:  "target",
		},

		"invalid-dir": {
			input: "/no-target/dir",
			want:  "fs",
		},

		"valid-base": {
			input: "/target",
			want:  "target",
		},

		"invalid-base": {
			input: "/no-target",
			want:  "fs",
		},

		"valid-base-slash": {
			input: "/target/",
			want:  "target",
		},

		"invalid-base-slash": {
			input: "/no-target/",
			want:  "fs",
		},

		"root": {
			input: "/",
			want:  "fs",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Target(test.input)

			if result != test.want {
				t.Errorf("unexpected regex match: %v, want %v", result, test.want)
			}
		})
	}
}

func TestIsValidRuleName(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"valid-lowercase": {
			input: "eicar",
			want:  true,
		},

		"valid-uppercase": {
			input: "EICAR",
			want:  true,
		},

		"valid-num": {
			input: "1",
			want:  true,
		},

		"valid-underscore": {
			input: "_",
			want:  true,
		},

		"valid-composite": {
			input: "1_aB",
			want:  true,
		},

		"invalid-space": {
			input: " ",
			want:  false,
		},

		"invalid-exclamation": {
			input: "!",
			want:  false,
		},

		"invalid-composite": {
			input: "a ",
			want:  false,
		},

		"invalid-empty": {
			input: "",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsValidYrName(test.input)

			if result != test.want {
				t.Errorf("unexpected rule name validation: %v, want %v", result, test.want)
			}
		})
	}
}

func TestFindRuleName(t *testing.T) {
	want := t.Name()

	if got := FindYrName(want); got != want {
		t.Errorf("unexpected rule name result: %v, want %v", got, want)
	}
}

func TestSetTargets(t *testing.T) {
	input := []*regexp.Regexp{
		regexp.MustCompile(t.Name()),
	}

	want := []*regexp.Regexp{
		targets[0],
		input[0],
	}

	if SetTargets(input...); !slices.Equal(targets, want) {
		t.Errorf("unexpected target result %v, want %v", targets, want)
	}
}
