// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/exec"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	New(env)
}

func TestLoad(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}

	if err = plat.Load(); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestExec(t *testing.T) {
	result, err := exec.Run("echo", t.Name())
	if err != nil {
		t.Fatalf("exec error: %v", err)
	}

	if string(result) != t.Name()+"\n" {
		t.Errorf("unexpected exec result %v, want %v", string(result), t.Name())
	}
}

func TestGetDomainInfo(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}

	want := []string{
		"/home/one/public_html",
		"/home/one/tmp",
		"/home/two/public_html",
		"/home/two/tmp",
		"/home/three/public_html",
		"/home/three/tmp",
	}

	result, err := plat.GetDocRoots()
	if err != nil {
		t.Fatalf("get domain error: %v", err)
	}

	if !slices.Equal(result, want) {
		t.Errorf("unexpected get domain info result %v, want %v", result, want)
	}
}

func TestGetDomainInfoErrs(t *testing.T) {
	mock := &Plat{
		bin: t.Name(),
	}

	if _, err := mock.GetDocRoots(); err != nil && !strings.HasPrefix(err.Error(), "exec: run error") {
		t.Errorf("unexpected get domain success")
	}
}

func TestCfg(t *testing.T) {
	var (
		plat = &Plat{
			cfg: &Cfg{},
		}

		got = plat.Cfg()
	)

	if !reflect.DeepEqual(got, plat.cfg) {
		t.Errorf("unexpected cfg result %v, want %v", got, plat.cfg)
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
		t.Errorf("unexpected acts result %v, want %v", got, plat.acters)
	}
}
