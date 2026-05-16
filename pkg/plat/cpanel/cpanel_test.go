// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/exec"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	New(env)
}

func TestLoad(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	if got := plat.Load(nil); got != nil {
		t.Errorf("load err %v", got)
	}
}

func TestExec(t *testing.T) {
	got, err := exec.Run("echo", t.Name())
	if err != nil {
		t.Fatalf("exec err %v", err)
	}

	if string(got) != t.Name()+"\n" {
		t.Errorf("unexpected exec result %v, want %v", string(got), t.Name())
	}
}

func TestDocRoots(t *testing.T) {
	var (
		input, err = Mock(t.Name(), t.TempDir())

		want = []string{
			"/home/one/public_html",
			"/home/one/tmp",
			"/home/two/public_html",
			"/home/two/tmp",
			"/home/three/public_html",
			"/home/three/tmp",
		}
	)

	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	got, err := input.DocRoots()
	if err != nil {
		t.Fatalf("docroots err %v", err)
	}

	if !slices.Equal(got, want) {
		t.Errorf("unexpected docroots result %v, want %v", got, want)
	}
}

func TestDocRootsErrs(t *testing.T) {
	input := &Plat{
		bin: t.Name(),
	}

	if _, got := input.DocRoots(); !errors.Is(got, exec.ErrRun) {
		t.Errorf("unexpected docroots success")
	}
}

func TestCfg(t *testing.T) {
	var (
		input = &Plat{
			cfg: &Cfg{},
		}

		got = input.Cfg()
	)

	if !reflect.DeepEqual(got, input.cfg) {
		t.Errorf("unexpected cfg result %v, want %v", got, input.cfg)
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
