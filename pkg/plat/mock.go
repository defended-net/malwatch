// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

type mock struct {
	acts []acter.Acter
}

type cfg struct{}

// Mock mocks a platform.
func Mock(acts ...acter.Acter) mock {
	return mock{
		acts: acts,
	}
}

// Load loads a given mocked plat.
func (plat mock) Load() error {
	return nil
}

// Acters returns a given mocked plat's active acts.
func (plat mock) Acters() []acter.Acter {
	return plat.acts
}

// Cfg returns a given mocked plat's cfg.
func (plat mock) Cfg() Cfg {
	return cfg{}
}

// Load loads a given mocked cfg.
func (cfg cfg) Load() error {
	return nil
}

// Path returns a given mocked cfg's path.
func (cfg cfg) Path() string {
	return ""
}
