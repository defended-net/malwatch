// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package json

import (
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	var (
		want = filepath.Join(t.TempDir(), t.Name())
		cfg  = NewCfg(want)
	)

	if cfg.Path() != want {
		t.Errorf("unexpected cfg path result %v, want %v", cfg.Path(), want)
	}
}
