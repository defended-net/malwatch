// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tbl

import (
	"reflect"
	"testing"
)

func TestPrepare(t *testing.T) {
	tests := map[string]struct {
		hdr   []string
		input [][]string
		want  struct {
			hdr  []string
			rows [][]string
		}
	}{
		"single": {
			hdr:   []string{"header"},
			input: [][]string{{"row"}},

			want: struct {
				hdr  []string
				rows [][]string
			}{
				hdr:  []string{"header"},
				rows: [][]string{{"row"}},
			},
		},

		"multiple": {
			hdr: []string{
				"header-a",
				"header-b",
			},

			input: [][]string{
				{
					"cell-a",
					"cell-b",
				},
				{
					"cell-c",
					"cell-d",
				},
			},

			want: struct {
				hdr  []string
				rows [][]string
			}{
				hdr: []string{
					"header-a",
					"header-b",
				},

				rows: [][]string{
					{"cell-a", "cell-b"},
					{"cell-c", "cell-d"},
				},
			},
		},

		"merge": {
			hdr: []string{
				"header",
			},

			input: [][]string{
				{
					"row-a",
					"row-a",
				},
			},
			want: struct {
				hdr  []string
				rows [][]string
			}{
				hdr:  []string{"header"},
				rows: [][]string{{"row-a\n"}},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hdr, rows := Prepare(test.hdr, test.input)

			if !reflect.DeepEqual(hdr, test.want.hdr) {
				t.Errorf("unexpected header %v, want %v", hdr, test.want.hdr)
			}

			if !reflect.DeepEqual(rows, test.want.rows) {
				t.Errorf("unexpected header %v, want %v", rows, test.want.rows)
			}
		})
	}
}

func TestDisplay(t *testing.T) {
	tests := map[string]struct {
		hdr   []string
		input [][]string
	}{
		"single": {
			hdr:   []string{"header"},
			input: [][]string{{"row"}},
		},

		"multiple": {
			hdr: []string{
				"header-a",
				"header-b",
			},

			input: [][]string{
				{
					"cell-a",
					"cell-b",
				},
				{
					"cell-c",
					"cell-d",
				},
			},
		},

		"merge": {
			hdr: []string{
				"header",
			},

			input: [][]string{
				{
					"row-a",
					"row-a",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hdr, rows := Prepare(test.hdr, test.input)

			if err := Print(t.Name(), hdr, rows); err != nil {
				t.Errorf("unexpected display error %v", err)
			}
		})
	}
}
