// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log/slog"
	"runtime/debug"

	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cmd"
)

func main() {
	var (
		state *cmd.State
		err   error
	)

	defer func() {
		if ok := recover(); ok != nil {
			slog.Error("panic", "stack", string(debug.Stack()))
		}

		cmd.Exit(state, err)
	}()

	state, err = cli.Run(cmds)
}
