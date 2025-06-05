// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"errors"
	"io/fs"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/cpanel"
	"github.com/defended-net/malwatch/pkg/plat/directadmin"
	"github.com/defended-net/malwatch/pkg/plat/preset"
)

// Load reads cfgs. Any plat cfgs not found will be written as file extension .disabled.
// The first plat cfg able to successfully load becomes the active plat.
func Load(env *env.Env) error {
	env.Plat = preset.New(env)

	plats := []plat.Plat{
		cpanel.New(env),
		directadmin.New(env),
	}

	for _, plat := range plats {
		err := fsys.InstallTOML(plat.Cfg().Path(), plat.Cfg())
		switch {
		case err == nil:
			continue

		case errors.Is(err, fs.ErrExist):
			env.Plat = plat

		default:
			return err
		}
	}

	return env.Plat.Load()
}
