// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package slack

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
		got   = NewCfg(input)

		want = &Cfg{
			path: input,
			User: "malwatch",
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	var (
		input = New(env)
		want  = acter.ErrDisabled
	)

	input.cfg.path = "slack.toml"

	if got := input.Load(env.Paths.Install.Root); got != nil &&
		!errors.Is(got, want) {
		t.Errorf("slack load err %v, want %v", got, want)
	}
}

func TestLoadErrs(t *testing.T) {
	var (
		sender = &Sender{
			cfg:     NewCfg("../escape"),
			secrets: &secret.Slack{},
		}

		input, err = os.OpenRoot(t.TempDir())
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = input.Close()
	}()

	if got := sender.Load(input); got == nil {
		t.Errorf("unexpected load success")
	}
}

func TestPath(t *testing.T) {
	var (
		want  = filepath.Join(t.TempDir(), t.Name())
		input = NewCfg(want)
	)

	if input.Path() != want {
		t.Errorf("unexpected cfg path result %v, want %v", input.Path(), want)
	}
}

func TestCfg(t *testing.T) {
	input, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	if got := input.Cfg(); got == nil {
		t.Errorf("unexpected cfg result")
	}
}
