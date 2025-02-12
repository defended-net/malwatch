// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package json

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestMain(m *testing.M) {
	if os.Getenv("JSON_ENDPOINT") == "" {
		fmt.Println("json: tests require env var")
		return

	}

	m.Run()
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := New(env).Load(); err != nil {
		t.Errorf("transport load error: %v", err)
	}
}

func TestNewAlert(t *testing.T) {
	tests := map[string]struct {
		input *state.Result
		want  error
	}{
		"hits": {
			input: &state.Result{},
			want:  nil,
		},
	}

	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer svc.Close()

	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock error: %v", err)
	}

	sender.cfg.Endpoint = svc.URL

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := sender.Alert(test.input); err != nil {
				t.Errorf("alert create error: %v", err)
			}
		})
	}
}

func TestSend(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer svc.Close()

	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock error: %v", err)
	}

	sender.cfg.Endpoint = svc.URL

	if err := sender.Alert(&state.Result{}); err != nil {
		t.Errorf("send error: %v", err)
	}
}

func TestCfgPath(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, "json.toml")

	env, err := env.Mock(t.Name(), dir)
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	got := New(env).Cfg().Path()

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
