// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package json

import (
	"errors"
	"io/fs"

	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents the cfg.
type Cfg struct {
	path     string
	Endpoint string
}

// NewCfg returns a cfg for given toml path.
func NewCfg(path string) *Cfg {
	return &Cfg{
		path:     path,
		Endpoint: "https://localhost",
	}
}

// Load reads the cfg from toml path.
func (cfg *Cfg) Load() error {
	if err := fsys.InstallTOML(cfg.path, cfg); !errors.Is(err, fs.ErrExist) {
		return err
	}

	return nil
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}
