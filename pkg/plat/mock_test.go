// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"testing"
)

func TestMockCfgLoad(t *testing.T) {
	mock := &mock{}

	if err := mock.Cfg().Load(); err != nil {
		t.Errorf("mock cfg load error: %v", err)
	}
}

func TestMockCfgPath(t *testing.T) {
	var (
		mock = &mock{}
		want = ""
	)

	if result := mock.Cfg().Path(); result != want {
		t.Errorf("unexpected cfg path result %v, want %v", result, want)
	}
}
