// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package history

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

// Cmd returns the cmd.
func Cmd() *cli.Cmd {
	return &cli.Cmd{
		Help: help.ErrHistory,
		Min:  1,

		Sub: cli.Sub{
			"get": {
				Help: help.ErrHistory,
				Fn:   Get,
			},

			"del": {
				Help: help.ErrHistory,
				Min:  1,
				Fn:   Del,
			},
		},
	}
}
