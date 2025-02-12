// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"errors"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/db"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	if err := Load(env); err != nil {
		t.Errorf("db load error: %v", err)
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

	if err := Load(env); err != nil {
		t.Errorf("db load error: %v", err)
	}
}

func TestLoadErrs(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"rel": {
			input: "home",
			want:  fsys.ErrPathNotAbs,
		},

		"not-exist": {
			input: "/dev/null",
			want:  fsys.ErrDirCreate,
		},
	}

	env := &env.Env{
		Cfg: &base.Cfg{
			Database: &db.Cfg{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env.Cfg.Database.Dir = test.input

			if err := Load(env); !errors.Is(err, test.want) {
				t.Errorf("db load error: %v", err)
			}
		})
	}
}
