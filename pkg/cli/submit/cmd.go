// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package submit

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

// Cmd returns the cmd.
func Cmd() *cli.Cmd {
	return &cli.Cmd{
		Help: help.ErrSubmit,
		Min:  1,
		Fn:   Do,
	}
}
