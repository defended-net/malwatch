// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cmd"
)

func main() {
	state, err := cli.Run(cmds)
	defer cmd.Exit(state, err)
}
