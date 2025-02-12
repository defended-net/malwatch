// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
)

func TestDo(t *testing.T) {
	env := &env.Env{
		Paths: &path.Paths{
			Install: &path.Install{
				Path: "",
			},
		},
	}

	if err := Do(env, []string{""}); err != nil {
		t.Errorf("install do error: %v", err)
	}
}

func TestDoNoArgs(t *testing.T) {
	if err := Do(nil, nil); err != nil {
		t.Errorf("install do error: %v", err)
	}
}
