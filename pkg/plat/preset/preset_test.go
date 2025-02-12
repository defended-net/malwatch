// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package preset

import (
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Plat = New(env)
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Plat = New(env)

	if err := env.Plat.Load(); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestCfg(t *testing.T) {
	var (
		want = &Cfg{}
		plat = &Plat{cfg: want}
		got  = plat.Cfg()
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}

func TestPath(t *testing.T) {
	var (
		want = t.TempDir()
		plat = &Plat{cfg: &Cfg{path: want}}
		got  = plat.Cfg().Path()
	)

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestActers(t *testing.T) {
	var (
		input = acter.Mock(act.VerbAlert)

		plat = &Plat{
			acters: []acter.Acter{
				input,
			},
		}

		got = plat.Acters()
	)

	if !reflect.DeepEqual(got, plat.acters) {
		t.Errorf("unexpected acts result %v, want %v", got, plat.acters)
	}
}
