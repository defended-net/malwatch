// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	client "github.com/defended-net/malwatch/pkg/client/http"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

var input = state.NewResult(
	"",

	state.Paths{
		"/target/test.php": {
			Rules:  []string{"eicar"},
			Status: "/quarantine/test.php-1",
		},
	},
)

func TestNewAlert(t *testing.T) {
	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock err %v", err)
	}

	if _, err := sender.NewAlert(input); err != nil {
		t.Errorf("new alert err %v", err)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		tmp      = t.TempDir()
		env, err = env.Mock(t.Name(), tmp)
		want     = filepath.Join(tmp, "pagerduty.toml")
	)

	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := New(env).Cfg().Path(); got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestCfgLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	input := New(env)
	input.cfg.path = "pagerduty.toml"

	if err := input.Load(env.Paths.Install.Root); err != nil {
		t.Errorf("sender cfg load err %v", err)
	}
}

func TestLoadDisabled(t *testing.T) {
	var (
		tmp        = t.TempDir()
		input, err = Mock(t.Name(), tmp)
		want       = acter.ErrDisabled
	)

	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	input.cfg.path = "pagerduty.toml"
	input.cfg.Endpoint = ""

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); !errors.Is(got, want) {
		t.Errorf("unexpected pagerduty load err %v, want %v", got, want)
	}
}

func TestAlert(t *testing.T) {
	tests := map[string]struct {
		status int
		want   error
	}{
		"ok": {
			status: http.StatusOK,
			want:   nil,
		},

		"accepted": {
			status: http.StatusAccepted,
			want:   nil,
		},

		"invalid-status": {
			status: http.StatusInternalServerError,
			want:   client.ErrBadStatus,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				wr.WriteHeader(test.status)
			}))
			defer svc.Close()

			sender, err := Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Fatalf("sender mock err %v", err)
			}

			sender.cfg.Endpoint = svc.URL

			if got := sender.Alert(input); !errors.Is(got, test.want) {
				t.Errorf("unexpected alert err %v, want %v", got, test.want)
			}
		})
	}
}

func TestAlertErrs(t *testing.T) {
	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("sender mock err %v", err)
	}

	sender.cfg.Endpoint = "http://127.0.0.1:0"

	if got := sender.Alert(input); !errors.Is(got, client.ErrReqDo) {
		t.Errorf("unexpected alert err %v, want %v", got, client.ErrReqDo)
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
