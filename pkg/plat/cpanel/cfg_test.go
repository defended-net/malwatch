// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCfgLoad(t *testing.T) {
	var (
		path = filepath.Join(t.TempDir(), t.Name())

		input = &Cfg{
			path: path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create error: %v", err)
	}

	if err := input.Load(); err != nil {
		t.Errorf("cfg load error: %v", err)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		want = t.Name()

		plat = &Plat{
			cfg: &Cfg{
				path: want,
			},
		}

		got = plat.cfg.Path()
	)

	if got != want {
		t.Errorf("unexpected cfg path %v, want %v", got, want)
	}
}
