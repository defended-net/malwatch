// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

func TestNewCfg(t *testing.T) {
	var (
		input = filepath.Join(t.TempDir(), t.Name())

		want = &Cfg{
			path: input,
		}
	)

	if got := NewCfg(input); !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected create smtp cfg result %v, want %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	input := New(env)
	input.cfg.path = "smtp.toml"

	got := input.Load(env.Paths.Install.Root)

	switch {
	case got == nil:

	case errors.Is(got, acter.ErrDisabled):

	default:
		t.Errorf("sender load err %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	input := &Sender{
		cfg:     NewCfg("../escape"),
		secrets: &secret.SMTP{},
	}

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); got == nil {
		t.Errorf("unexpected load success")
	}
}

func TestPath(t *testing.T) {
	var (
		tmp      = t.TempDir()
		env, err = env.Mock(t.Name(), tmp)
		want     = filepath.Join(tmp, "smtp.toml")
	)

	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := New(env).Cfg().Path(); got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
