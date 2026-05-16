// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestLoad(t *testing.T) {
	var (
		root, err = os.OpenRoot(t.TempDir())
		input     = NewCfg(t.Name())
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); got != nil {
		t.Errorf("cfg load err %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	var (
		root, err = os.OpenRoot(t.TempDir())
		input     = NewCfg("../escape")
		want      = fsys.ErrPathLocal
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); !errors.Is(got, want) {
		t.Errorf("unexpected cfg load err %v, want %v", got, want)
	}
}

func TestLoadExist(t *testing.T) {
	var (
		tmp       = t.TempDir()
		name      = "cfg.toml"
		path      = filepath.Join(tmp, name)
		root, err = os.OpenRoot(tmp)
		input     = NewCfg(name)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if got := input.Load(root); got != nil {
		t.Errorf("unexpected cfg load err %v, want nil", got)
	}
}

func TestPath(t *testing.T) {
	var (
		want = t.Name()
		got  = NewCfg(want)
	)

	if got.Path() != want {
		t.Errorf("unexpected cfg path result %v, want %v", got.Path(), want)
	}
}
