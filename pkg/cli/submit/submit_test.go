// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package submit

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestDo(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %s", err)
	}

	path := filepath.Join(t.TempDir(), t.Name())

	file, err := os.Create(path)
	if err != nil {
		t.Fatal("file create error:", err)
	}
	defer file.Close()

	if err := Do(env, []string{path}); err != nil {
		t.Errorf("submit error: %v", err)
	}
}

func TestDoInvalidPath(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"invalid": {
			input: []string{
				"\\",
			},

			want: fsys.ErrPathNotAbs,
		},

		"rel": {
			input: []string{
				"target/index.php",
			},

			want: fsys.ErrPathNotAbs,
		},

		"file": {
			input: []string{
				"index.php",
			},

			want: fsys.ErrPathNotAbs,
		},

		"space": {
			input: []string{
				" ",
			},

			want: fsys.ErrPathNotAbs,
		},

		"none": {
			input: []string{
				"",
			},

			want: fsys.ErrPathNotAbs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Errorf("env mock error: %v", err)
			}

			if err := Do(env, test.input); !errors.Is(err, test.want) {
				t.Errorf("unexpected submit result %v, want %v", err, test.want)
			}
		})
	}
}

func TestDoErrs(t *testing.T) {
	var (
		input = []string{filepath.Join(t.TempDir(), t.Name())}
		want  = fs.ErrNotExist
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := Do(env, input); !errors.Is(err, want) {
		t.Errorf("unexpected submit err %v, want %v", err, want)
	}
}
