// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"path/filepath"
	"testing"
)

func TestMock(t *testing.T) {
	var (
		want = filepath.Join(t.TempDir(), t.Name())
		got  = Mock(want).Dir
	)

	if got != want {
		t.Errorf("unexpected logger mock result %v, want %v", got, want)
	}
}
