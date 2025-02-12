// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents the cfg.
type Cfg struct {
	path     string
	SkipAccs []string
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}

// Load loads the cfg.
func (cfg *Cfg) Load() error {
	return fsys.ReadTOML(cfg.Path(), cfg)
}
