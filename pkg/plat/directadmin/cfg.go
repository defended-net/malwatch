// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import (
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents the cfg.
type Cfg struct {
	path     string
	User     string
	SkipAccs []string
}

// Load loads the cfg.
func (cfg *Cfg) Load() error {
	if err := fsys.ReadTOML(cfg.Path(), cfg); err != nil {
		return err
	}

	if cfg.User == "" {
		cfg.User = "admin"
	}

	return nil
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}
