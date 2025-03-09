// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package quarantine

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
	"github.com/defended-net/malwatch/pkg/sig"
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

	if _, err := os.Create(path); err != nil {
		t.Errorf("file create error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Errorf("sigs mock error: %v", err)
	}

	env.Plat = plat.Mock(acter.Mock(act.VerbQuarantine))

	if err := Do(env, []string{path}); err != nil {
		t.Errorf("quarantine error: %v", err)
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

		"none": {
			input: []string{
				"",
			},

			want: fsys.ErrPathNotAbs,
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Do(env, test.input); !errors.Is(err, test.want) {
				t.Errorf("unexpected quarantine result %v, want %v", err, test.want)
			}
		})
	}
}

func TestDoNotExist(t *testing.T) {
	var (
		input = []string{filepath.Join(t.TempDir(), t.Name())}
		want  = fs.ErrNotExist
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := Do(env, input); !errors.Is(err, want) {
		t.Errorf("unexpected quarantine err %v, want %v", err, want)
	}
}
