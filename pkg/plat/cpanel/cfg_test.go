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
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())

		input = &Cfg{
			path: path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create err %v", err)
	}

	root, err := os.OpenRoot(tmp)
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
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
