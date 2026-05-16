// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"testing"
)

func TestIsSafeExpr(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"empty": {
			input: "",
			want:  ErrCleanExprNotSafe,
		},

		"sub": {
			input: `s/foo/bar/g`,
			want:  nil,
		},

		"trans": {
			input: `y/foo/bar/`,
			want:  nil,
		},

		"del": {
			input: `d`,
			want:  nil,
		},

		"label": {
			input: `:label`,
			want:  nil,
		},

		"branch": {
			input: `b label`,
			want:  nil,
		},

		"comment": {
			input: "# comment\nd",
			want:  nil,
		},

		"semi": {
			input: `;`,
			want:  nil,
		},

		"newline": {
			input: "d\nd",
			want:  nil,
		},

		"hold": {
			input: `g`,
			want:  nil,
		},

		"app": {
			input: `a\\txt`,
			want:  nil,
		},

		"quit": {
			input: `q`,
			want:  nil,
		},

		"addr": {
			input: `1d`,
			want:  nil,
		},

		"blk": {
			input: `{d}`,
			want:  nil,
		},

		"negate": {
			input: `!d`,
			want:  nil,
		},

		"re-addr": {
			input: `/pattern/d`,
			want:  nil,
		},

		"unsupp-exec": {
			input: `e`,
			want:  ErrCleanSuppExec,
		},

		"unsupp-write": {
			input: `w file`,
			want:  ErrCleanSuppWrite,
		},

		"unsupp-read": {
			input: `r file`,
			want:  ErrCleanSuppRead,
		},

		"unsafe-flag": {
			input: `s/foo/bar/we`,
			want:  ErrCleanExprNotSafe,
		},

		"unsafe-flag-w": {
			input: `s/foo/bar/wfile`,
			want:  ErrCleanExprNotSafe,
		},

		"unsafe-addr-range": {
			input: `1,5d`,
			want:  ErrCleanExprNotSafe,
		},

		"unsafe": {
			input: `Z`,
			want:  ErrCleanExprNotSafe,
		},

		"unsafe-re": {
			input: `/pattern`,
			want:  ErrCleanExprNotSafe,
		},

		"unsafe-sub": {
			input: `s/foo`,
			want:  ErrCleanExprNotSafe,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := isSafeExpr(test.input)

			switch {
			case test.want == nil && got != nil:
				t.Errorf("is safe expr err %v", got)

			case test.want != nil && !errors.Is(got, test.want):
				t.Errorf("unexpected is safe expr err %v, want %v", got, test.want)
			}
		})
	}
}
