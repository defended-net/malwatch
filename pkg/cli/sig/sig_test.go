// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestUpdate(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := Update(env, []string{}); got != nil {
		t.Errorf("update err %v", got)
	}
}

func TestRefresh(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := Refresh(env, []string{""}); got != nil {
		t.Errorf("refresh err %v", got)
	}
}
