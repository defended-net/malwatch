// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCfgLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create error: %v", err)
	}

	var (
		got = &Cfg{
			path: path,
		}

		want = &Cfg{
			path: path,
			User: "admin",
		}
	)

	if err := got.Load(); err != nil {
		t.Fatalf("cfg load error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg load result %v, want %v", got, nil)
	}
}

func TestCfgLoadCustomUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create error: %v", err)
	}

	var (
		got = &Cfg{
			path: path,
			User: t.Name(),
		}

		want = &Cfg{
			path: path,
			User: t.Name(),
		}
	)

	if err := got.Load(); err != nil {
		t.Fatalf("cfg load error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg load result %v, want %v", got, nil)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		want = filepath.Join(t.TempDir(), t.Name())

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
