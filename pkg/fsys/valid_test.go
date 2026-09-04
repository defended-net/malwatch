// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"path/filepath"
	"testing"
)

func TestIsUnder(t *testing.T) {
	tests := map[string]struct {
		input string
		path  string
		want  bool
	}{
		"match": {
			input: "/test",
			path:  "/test",
			want:  true,
		},

		"root": {
			input: "/test/root",
			path:  "/",
			want:  true,
		},

		"rel": {
			input: "/a",
			path:  "./b/c",
			want:  false,
		},

		"parent": {
			input: "/test",
			path:  "/test/parent",
			want:  false,
		},

		"child": {
			input: "/test/child",
			path:  "/test",
			want:  true,
		},

		"nested": {
			input: "/test/a/b/nested",
			path:  "/test",
			want:  true,
		},

		"sibling-pfx": {
			input: "/test/a/sibling",
			path:  "/test/b",
			want:  false,
		},

		"trail-sep": {
			input: "/test/trail",
			path:  "/test/",
			want:  true,
		},

		"zero-len": {
			input: "/test/a",
			path:  "",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsUnder(test.input, test.path); got != test.want {
				t.Fatalf("unexpected is under result %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsUnderBases(t *testing.T) {
	tests := map[string]struct {
		input string
		bases []string
		want  bool
	}{
		"is-under": {
			input: filepath.Join(t.Name(), "/base/index.php"),

			bases: []string{
				filepath.Join(t.Name(), "/base/config"),
				filepath.Join(t.Name(), "/base/assets"),
				filepath.Join(t.Name(), "/base"),
			},

			want: true,
		},

		"not-under": {
			input: filepath.Join(t.Name(), "/base/index.php"),

			bases: []string{
				"/base",
				"/base-b",
			},

			want: false,
		},

		"empty": {
			input: filepath.Join(t.Name(), "/base/index.php"),
			bases: []string{},
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsUnder(test.input, test.bases...); got != test.want {
				t.Fatalf("unexpected is under result %v, want %v", got, test.want)
			}
		})
	}
}
