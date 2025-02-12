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
		t.Errorf("env mock error: %v", err)
	}

	if err := Update(env, []string{}); err != nil {
		t.Errorf("update error: %v", err)
	}
}

func TestRefresh(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := Refresh(env, []string{""}); err != nil {
		t.Errorf("refresh error: %v", err)
	}
}
