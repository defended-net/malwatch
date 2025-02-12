// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

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
		input = &Cfg{
			path: path,
		}

		got = input.Load()
	)

	if !reflect.DeepEqual(got, nil) {
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

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}
