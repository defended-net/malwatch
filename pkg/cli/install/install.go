// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/install"
)

// Do helps avoid error alarms by creating a clean 0 exit status code for initial run.
func Do(env *env.Env, args []string) error {
	if len(args) == 0 {
		return nil
	}

	return install.Sysd("/etc/systemd/system", env.Paths.Install.Path+"-monitor")
}
