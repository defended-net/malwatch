// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package preset

import (
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

// Plat represents the default platform.
type Plat struct {
	env    *env.Env
	cfg    *Cfg
	acters []acter.Acter
}

// Cfg represents the cfg.
type Cfg struct {
	path string
}

// New returns a plat from given env.
func New(env *env.Env) *Plat {
	return &Plat{
		env:    env,
		acters: act.Preset(env),
		cfg:    &Cfg{},
	}
}

// Load reads given plat cfg files.
func (plat *Plat) Load() error {
	acters, err := acter.Load(plat.acters)
	if err != nil {
		return err
	}

	plat.acters = acters

	return nil
}

// Acters returns given plat's acters.
func (plat *Plat) Acters() []acter.Acter {
	return plat.acters
}

// Cfg returns given plat's cfg.
func (plat *Plat) Cfg() plat.Cfg {
	return plat.cfg
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}

// Load loads the cfg.
func (cfg *Cfg) Load() error {
	return nil
}
