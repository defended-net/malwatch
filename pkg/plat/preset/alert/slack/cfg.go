// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package slack

import (
	"errors"
	"io/fs"
	"os"

	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents cfg.
type Cfg struct {
	path    string
	User    string
	Channel string
}

// NewCfg returns cfg for given toml path.
func NewCfg(path string) *Cfg {
	return &Cfg{
		path: path,
		User: "malwatch",
	}
}

// Load reads cfg from toml path.
func (cfg *Cfg) Load(root *os.Root) error {
	if err := fsys.InstallTOML(root, cfg.path, cfg); !errors.Is(err, fs.ErrExist) {
		return err
	}

	return nil
}

// Path returns given cfg toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}
