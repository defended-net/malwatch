// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/db"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	if got := Load(env); got != nil {
		t.Errorf("db load err %v", got)
	}
}

func TestLoadNoDb(t *testing.T) {
	env := &env.Env{
		Cfg: &base.Cfg{
			Database: &db.Cfg{
				Dir: "",
			},
		},
	}

	if got := Load(env); got != nil {
		t.Errorf("db load err %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	var (
		tmp     = t.TempDir()
		blocker = filepath.Join(tmp, "blocker")

		tests = map[string]struct {
			dir  string
			want error
		}{
			"rel": {
				dir:  "home",
				want: fsys.ErrPathNotAbs,
			},

			"not-local": {
				dir:  "/not-local",
				want: fsys.ErrPathLocal,
			},

			"mkdir": {
				dir:  filepath.Join(blocker, "child"),
				want: fsys.ErrDirCreate,
			},
		}
	)

	if _, err := os.Create(blocker); err != nil {
		t.Fatalf("file create err %v", err)
	}

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := &env.Env{
				Cfg: &base.Cfg{
					Database: &db.Cfg{
						Dir: test.dir,
					},
				},

				Paths: &path.Paths{
					Install: &path.Install{
						Root: root,
						Db:   filepath.Join(test.dir, "x.db"),
					},
				},
			}

			if got := Load(env); !errors.Is(got, test.want) {
				t.Errorf("db load err %v, want %v", got, test.want)
			}
		})
	}
}
