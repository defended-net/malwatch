// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestDo(t *testing.T) {
	path := filepath.Join(t.TempDir(), t.Name())

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %s", err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatal("file create err", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	env.Plat = plat.Mock(acter.Mock(act.VerbExile, true))

	if got := Do(env, []string{path}); got != nil {
		t.Errorf("exile err %v", got)
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
				t.Errorf("env mock err %v", err)
			}

			env.Cfg.Acts.Quarantine.Dir = t.TempDir()

			if got := Do(env, test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected do result %v, want %v", got, test.want)
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
		t.Errorf("env mock err %v", err)
	}

	if got := Do(env, input); !errors.Is(got, want) {
		t.Errorf("unexpected do err %v, want %v", got, want)
	}
}
