// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
	"github.com/defended-net/malwatch/pkg/cli/history"
	"github.com/defended-net/malwatch/pkg/cli/monitor"
)

// https://developers.google.com/style/code-syntax

// cmds stores arg layout specific to cmd.
var cmds = cli.Layout{
	"start": {
		Help: help.ErrMonitor,
		Min:  0,
		Fn:   monitor.Do,
	},

	"history": {
		Help: help.ErrHistory,
		Min:  1,

		Layout: cli.Layout{
			"get": {
				Help: help.ErrHistory,
				Min:  0,
				Fn:   history.Get,
			},

			"del": {
				Help: help.ErrHistory,
				Min:  1,
				Fn:   history.Del,
			},
		},
	},
}
