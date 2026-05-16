// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestAlert(t *testing.T) {
	if os.Getenv("SMTP_HOSTNAME") == "" {
		t.Skip("smtp: requires SMTP_HOSTNAME env var")
	}

	input, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock err %v", err)
	}

	if got := input.Alert(&state.Result{}); got != nil {
		t.Errorf("alert err %v", got)
	}
}

func TestLoadDisabled(t *testing.T) {
	var (
		env, err = env.Mock(t.Name(), t.TempDir())
		want     = acter.ErrDisabled
	)

	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Cfg.Secrets.Alerts.SMTP.Hostname = ""

	input := New(env)
	input.cfg.path = "smtp.toml"

	if got := input.Load(env.Paths.Install.Root); !errors.Is(got, want) {
		t.Errorf("unexpected sender load err %v, want %v", got, want)
	}
}

func TestSenderLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Cfg.Secrets.Alerts.SMTP.Hostname = "smtp.example.com"

	input := New(env)
	input.cfg.path = "smtp.toml"

	if got := input.Load(env.Paths.Install.Root); got != nil {
		t.Errorf("sender load err %v", got)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		tmp      = t.TempDir()
		env, err = env.Mock(t.Name(), tmp)
		want     = filepath.Join(tmp, "smtp.toml")
	)

	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := New(env).Cfg().Path(); got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestAlertInvalidFrom(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	input := New(env)
	input.cfg.From = "invalid"

	if got := input.Alert(&state.Result{}); got == nil {
		t.Errorf("unexpected alert success")
	}
}

func TestAlertInvalidTo(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	sender := New(env)
	sender.cfg.From = "from@example.com"
	sender.cfg.To = []string{"invalid"}

	if got := sender.Alert(&state.Result{}); got == nil {
		t.Errorf("unexpected alert success")
	}
}

func TestMock(t *testing.T) {
	t.Setenv("SMTP_HOSTNAME", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")

	if _, got := Mock(t.Name(), t.TempDir()); got != nil {
		t.Errorf("sender mock err %v", got)
	}
}
