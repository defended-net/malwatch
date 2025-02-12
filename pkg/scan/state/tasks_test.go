// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/db/orm"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
)

func TestPrint(t *testing.T) {
	tests := map[string]struct {
		input *Result
		want  []error
	}{
		"single": {
			input: &Result{
				Target: "target",

				Paths: map[string]*hit.Meta{
					"/target/index.php": {
						Rules: []string{"rule"},
					},
				},
			},
		},

		"compound": {
			input: &Result{
				Target: "target",

				Paths: map[string]*hit.Meta{
					"/target/index.php": {
						Rules: []string{"rule"},
					},

					"/target/index-b.php": {
						Rules: []string{"rule-b"},
					},

					"/target/index-c.php": {
						Rules: []string{"rule-c"},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.input.Print(); err != nil {
				t.Errorf("print error: %v", err)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := map[string]struct {
		input *Result
		want  []error
	}{
		"single": {
			input: &Result{
				Target: "target",

				Paths: map[string]*hit.Meta{
					"/target/index.php": {
						Rules: []string{"rule"},
					},
				},

				Errs: &Errs{},
			},
		},

		"compound": {
			input: &Result{
				Target: "target",

				Paths: map[string]*hit.Meta{
					"/target/index.php": {
						Rules: []string{"rule"},
					},

					"/target/index-b.php": {
						Rules: []string{"rule-b"},
					},

					"/target/index-c.php": {
						Rules: []string{"rule-c"},
					},
				},

				Errs: &Errs{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
			if err != nil {
				t.Fatalf("db mock error: %v", err)
			}

			if err := test.input.Save(db); err != nil {
				t.Errorf("db save error: %v", err)
			}
		})
	}
}

func TestLog(t *testing.T) {
	input := &Result{}

	if got := input.Log(); got != nil {
		t.Errorf("unexpected log err %v, want %v", got, nil)
	}
}
