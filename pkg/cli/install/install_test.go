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

	if got := Do(env, []string{""}); got != nil {
		t.Errorf("install do err %v", got)
	}
}

func TestDoNoArgs(t *testing.T) {
	if got := Do(nil, nil); got != nil {
		t.Errorf("install do err %v", got)
	}
}
