// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"os"

	"github.com/defended-net/malwatch/pkg/plat/acter"
)

// Plat represents the platform.
type Plat interface {
	Load(*os.Root) error
	Cfg() Cfg
	Acters() []acter.Acter
}

// Cfg represents a cfg.
type Cfg interface {
	Load(*os.Root) error
	Path() string
}
