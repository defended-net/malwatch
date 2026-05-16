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
		t.Fatalf("env mock err %v", err)
	}

	env.Plat = New(env)
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if env.Cfg.Secrets.S3.Endpoint == "" {
		t.Skip("s3 secret not stored")
	}

	env.Cfg.Acts.Quarantine.Dir = ""

	env.Plat = New(env)

	if got := env.Plat.Load(env.Paths.Install.Root); err != got {
		t.Errorf("plat load err %v", got)
	}
}

func TestCfg(t *testing.T) {
	var (
		want = &Cfg{}

		input = &Plat{
			cfg: want,
		}

		got = input.Cfg()
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}

func TestPath(t *testing.T) {
	var (
		want = t.TempDir()

		input = &Plat{
			cfg: &Cfg{
				path: want,
			},
		}

		got = input.Cfg().Path()
	)

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestActers(t *testing.T) {
	var (
		input = acter.Mock(act.VerbAlert, true)

		plat = &Plat{
			acters: []acter.Acter{
				input,
			},
		}

		got = plat.Acters()
	)

	if !reflect.DeepEqual(got, plat.acters) {
		t.Errorf("unexpected acters result %v, want %v", got, plat.acters)
	}
}
