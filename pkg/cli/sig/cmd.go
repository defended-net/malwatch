// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

// Cmd returns the cmd.
func Cmd() *cli.Cmd {
	return &cli.Cmd{
		Help: help.ErrSigs,
		Min:  1,

		Sub: cli.Sub{
			"update": {
				Help: help.ErrSigs,
				Fn:   Update,
			},

			"refresh": {
				Help: help.ErrSigs,
				Fn:   Refresh,
			},
		},
	}
}
