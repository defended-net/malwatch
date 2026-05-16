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
		got  = NewCfg(want)
	)

	if got.Path() != want {
		t.Errorf("unexpected cfg path result %v, want %v", got.Path(), want)
	}
}
