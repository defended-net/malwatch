// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"os"
	"testing"
)

func TestMockLoad(t *testing.T) {
	input := &mock{}

	if got := input.Load(nil); got != nil {
		t.Errorf("mock load err %v", got)
	}
}

func TestMockCfgLoad(t *testing.T) {
	var (
		input     = &mock{}
		root, err = os.OpenRoot(t.TempDir())
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Cfg().Load(root); got != nil {
		t.Errorf("mock cfg load err %v", got)
	}
}

func TestMockCfgPath(t *testing.T) {
	var (
		input = &mock{}
		want  = ""
	)

	if got := input.Cfg().Path(); got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}
