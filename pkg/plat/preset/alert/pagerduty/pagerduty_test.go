// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

var (
	hits = &state.Result{
		Paths: map[string]*hit.Meta{
			"/target/test.php": {
				Rules:  []string{"eicar"},
				Status: "/quarantine/test.php-1",
			},
		},
	}
)

func TestMain(m *testing.M) {
	if os.Getenv("PD_TOKEN") == "" {
		fmt.Println("pagerduty: tests require env var")
		return

	}

	m.Run()
}

func TestNewAlert(t *testing.T) {
	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock error: %v", err)
	}

	if _, err := sender.NewAlert(hits); err != nil {
		t.Errorf("new alert error: %v", err)
	}
}

func TestCfgPath(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, "pagerduty.toml")

	env, err := env.Mock(t.Name(), dir)
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	got := New(env).Cfg().Path()

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestCfgLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	sender := New(env)

	if err := sender.Load(); err != nil {
		t.Errorf("sender cfg load error: %v", err)
	}
}
