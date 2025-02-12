// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

func TestNewCfg(t *testing.T) {
	var (
		input = filepath.Join(t.TempDir(), t.Name())

		want = &Cfg{
			path: input,
		}

		got = NewCfg(input)
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if got := New(env).Load(); got != nil {
		t.Errorf("sender load error: %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	sender := &Sender{
		cfg:     NewCfg("/dev/null/not-exist"),
		secrets: &secret.SMTP{},
	}

	if got := sender.Load(); got == nil {
		t.Errorf("unexpected load success: %v", got)
	}
}

func TestPath(t *testing.T) {
	var (
		dir  = t.TempDir()
		want = filepath.Join(dir, "smtp.toml")
	)

	env, err := env.Mock(t.Name(), dir)
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	got := New(env).Cfg().Path()

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
