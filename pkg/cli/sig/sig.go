// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/sig"
)

// Update does a sig update.
func Update(env *env.Env, _ []string) error {
	return sig.Update(env)
}

// Refresh does a sig refresh.
func Refresh(env *env.Env, _ []string) error {
	return sig.Refresh(env)
}
