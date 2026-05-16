// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package json

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	input := New(env)
	input.cfg.path = "json.toml"

	if got := input.Load(env.Paths.Install.Root); err != got {
		t.Errorf("transport load err %v", got)
	}
}

func TestNewAlert(t *testing.T) {
	var (
		svc        = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		input, err = Mock(t.Name(), t.TempDir())

		tests = map[string]struct {
			input *state.Result
			want  error
		}{
			"hits": {
				input: &state.Result{},
				want:  nil,
			},
		}
	)

	defer svc.Close()

	if err != nil {
		t.Errorf("sender mock err %v", err)
	}

	input.cfg.Endpoint = svc.URL

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := input.Alert(test.input); got != nil {
				t.Errorf("create alert err %v", got)
			}
		})
	}
}

func TestSend(t *testing.T) {
	var (
		svc        = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		input, err = Mock(t.Name(), t.TempDir())
	)

	defer svc.Close()

	if err != nil {
		t.Errorf("sender mock err %v", err)
	}

	input.cfg.Endpoint = svc.URL

	if got := input.Alert(&state.Result{}); got != nil {
		t.Errorf("send err %v", got)
	}
}

func TestLoadDisabled(t *testing.T) {
	var (
		tmp         = t.TempDir()
		sender, err = Mock(t.Name(), tmp)
		want        = acter.ErrDisabled
	)

	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	sender.cfg.path = "json.toml"
	sender.cfg.Endpoint = ""

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := sender.Load(root); !errors.Is(got, want) {
		t.Errorf("unexpected load err %v, want %v", got, want)
	}
}

func TestAlertErrs(t *testing.T) {
	input, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	input.cfg.Endpoint = "http://127.0.0.1:0"

	if got := input.Alert(&state.Result{}); got == nil {
		t.Errorf("unexpected alert success")
	}
}

func TestCfg(t *testing.T) {
	input, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	if got := input.Cfg(); got == nil {
		t.Errorf("unexpected cfg result %v", got)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		tmp      = t.TempDir()
		env, err = env.Mock(t.Name(), tmp)
		want     = filepath.Join(tmp, "json.toml")
	)

	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := New(env).Cfg().Path(); got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
