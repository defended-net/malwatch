// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestLoad(t *testing.T) {
	cfg := NewCfg(filepath.Join(t.TempDir(), t.Name()))

	if err := cfg.Load(); err != nil {
		t.Errorf("cfg load error: %v", err)
	}
}

func TestLoadErrs(t *testing.T) {
	cfg := NewCfg(filepath.Join("/dev/null", t.Name()))

	if err := cfg.Load(); !errors.Is(err, fsys.ErrTOMLRead) {
		t.Errorf("unexpected cfg load error: %v, want %v", err, fsys.ErrTOMLRead)
	}
}

func TestPath(t *testing.T) {
	want := filepath.Join(t.TempDir(), t.Name())

	cfg := NewCfg(want)

	if cfg.Path() != want {
		t.Errorf("unexpected cfg path result %v, want %v", cfg.Path(), want)
	}
}
