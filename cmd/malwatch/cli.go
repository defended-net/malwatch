// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/act"
	"github.com/defended-net/malwatch/pkg/cli/exile"
	"github.com/defended-net/malwatch/pkg/cli/history"
	"github.com/defended-net/malwatch/pkg/cli/info"
	"github.com/defended-net/malwatch/pkg/cli/install"
	"github.com/defended-net/malwatch/pkg/cli/quarantine"
	"github.com/defended-net/malwatch/pkg/cli/restore"
	"github.com/defended-net/malwatch/pkg/cli/scan"
	"github.com/defended-net/malwatch/pkg/cli/sig"
	"github.com/defended-net/malwatch/pkg/cli/submit"
)

// cmds stores arg:cmd layout.
var cmds = cli.Sub{
	"scan":       scan.Cmd(),
	"history":    history.Cmd(),
	"actions":    act.Cmd(),
	"quarantine": quarantine.Cmd(),
	"exile":      exile.Cmd(),
	"restore":    restore.Cmd(),
	"submit":     submit.Cmd(),
	"signatures": sig.Cmd(),
	"info":       info.Cmd(),
	"install":    install.Cmd(),
}
