// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/act"
	"github.com/defended-net/malwatch/pkg/cli/exile"
	"github.com/defended-net/malwatch/pkg/cli/help"
	"github.com/defended-net/malwatch/pkg/cli/history"
	"github.com/defended-net/malwatch/pkg/cli/info"
	"github.com/defended-net/malwatch/pkg/cli/install"
	"github.com/defended-net/malwatch/pkg/cli/quarantine"
	"github.com/defended-net/malwatch/pkg/cli/restore"
	"github.com/defended-net/malwatch/pkg/cli/scan"
	"github.com/defended-net/malwatch/pkg/cli/sig"
)

// https://developers.google.com/style/code-syntax

// cmds stores arg layout specific to cmd.
var cmds = cli.Layout{
	"scan": {
		Help: help.ErrScan,
		Min:  0,
		Fn:   scan.Do,
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

	"actions": {
		Help: help.ErrActs,
		Min:  1,

		Layout: cli.Layout{
			"get": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   act.Get,
			},

			"set": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   act.Set,
			},

			"del": {
				Help: help.ErrActs,
				Min:  1,
				Fn:   act.Del,
			},
		},
	},

	"quarantine": {
		Help: help.ErrQuarantine,
		Min:  1,
		Fn:   quarantine.Do,
	},

	"exile": {
		Help: help.ErrExile,
		Min:  1,
		Fn:   exile.Do,
	},

	"restore": {
		Help: help.ErrRestore,
		Min:  1,
		Fn:   restore.Do,
	},

	"signatures": {
		Help: help.ErrSigs,
		Min:  1,

		Layout: cli.Layout{
			"update": {
				Help: help.ErrSigs,
				Min:  0,
				Fn:   sig.Update,
			},

			"refresh": {
				Help: help.ErrSigs,
				Min:  0,
				Fn:   sig.Refresh,
			},
		},
	},

	"info": {
		Help: help.ErrInfo,
		Min:  0,
		Fn:   info.Do,
	},

	"install": {
		Help: help.ErrInstall,
		Min:  0,
		Fn:   install.Do,
	},
}
