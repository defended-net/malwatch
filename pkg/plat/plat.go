// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

// Plat represents the platform.
type Plat interface {
	Load() error
	Cfg() Cfg
	Acters() []acter.Acter
}

// Cfg represents a cfg.
type Cfg interface {
	Load() error
	Path() string
}
