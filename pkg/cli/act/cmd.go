// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

// Cmd returns the cmd.
func Cmd() *cli.Cmd {
	return &cli.Cmd{
		Help: help.ErrActs,
		Min:  1,

		Sub: cli.Sub{
			"get": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   Get,
			},

			"set": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   Set,
			},

			"del": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   Del,
			},
		},
	}
}
