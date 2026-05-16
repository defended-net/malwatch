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
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())
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

	var (
		got = &Cfg{
			path: path,
		}

		want = &Cfg{
			path: path,
			User: "admin",
		}
	)

	if err := got.Load(root); err != nil {
		t.Fatalf("cfg load err %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg load result %v, want %v", got, nil)
	}
}

func TestCfgLoadCustomUser(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())
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

	if err := got.Load(root); err != nil {
		t.Fatalf("cfg load err %v", err)
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
