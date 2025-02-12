// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"testing"

	"github.com/defended-net/malwatch/third_party/yr"
)

func TestHasMatch(t *testing.T) {
	tests := map[string]struct {
		matches *yr.MatchRules
		rule    string
		want    bool
	}{
		"single-match": {
			matches: &yr.MatchRules{
				{
					Rule: "eicar",
				},
			},

			rule: "eicar",
			want: true,
		},

		"single-no-match": {
			matches: &yr.MatchRules{
				{
					Rule: "eicar",
				},
			},

			rule: "rule",
			want: false,
		},

		"multiple-match": {
			matches: &yr.MatchRules{
				{
					Rule: "rule",
				},
				{
					Rule: "eicar",
				},
			},

			rule: "eicar",
			want: true,
		},

		"multiple-no-match": {
			matches: &yr.MatchRules{
				{
					Rule: "rule",
				},
				{
					Rule: "eicar",
				},
			},

			rule: "no-match",
			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := HasMatch(test.matches, test.rule); got != test.want {
				t.Errorf("unexpected has match result %v, want %v", got, test.want)
			}
		})
	}
}
