// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/monitor"
)

// cmds stores arg:cmd layout.
var cmds = cli.Sub{
	"start": monitor.Cmd(),
}
