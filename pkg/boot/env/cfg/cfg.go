// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cfg

import (
	"path/filepath"

	vars "github.com/caarlos0/env/v11"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

// Cfg represents a cfg.
type Cfg interface {
	Load() error
	Path() string
}

// Load reads all cfgs. Path definitions are loaded according to user defined cfg values.
func Load(env *env.Env) error {
	env.Cfg = base.New(env.Paths)
	env.Cfg.Acts = act.New(env.Paths.Cfg.Acts)
	env.Cfg.Secrets = secret.New(env.Paths.Cfg.Secrets)

	cfgs := []Cfg{
		env.Cfg,
		env.Cfg.Acts,
		env.Cfg.Secrets,
	}

	for _, cfg := range cfgs {
		if err := cfg.Load(); err != nil {
			return err
		}
	}

	// Log and db files are relative to binary name. Avoids concurrent access complexity.
	env.Paths.Install.Log = filepath.Join(env.Cfg.Log.Dir, env.Paths.Install.Bin+".log")
	env.Paths.Install.Db = filepath.Join(env.Cfg.Database.Dir, "malwatch.db")

	return vars.Parse(env.Cfg)
}
