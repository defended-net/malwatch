// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"flag"
	"fmt"

	"github.com/defended-net/malwatch/pkg/boot"
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/cli/help"
	"github.com/defended-net/malwatch/pkg/cmd"
)

// Layout represents cmd layout.
type Layout map[string]*Cmd

// Cmd represents a cmd.
type Cmd struct {
	Help   error
	Args   []string
	Min    int
	Pos    int
	Layout Layout
	Fn     CmdFn
}

// CmdFn represents a cmd fn.
type CmdFn func(*env.Env, []string) error

// Run boots, locks and starts the cmd.
func Run(layout Layout) (*cmd.State, error) {
	state := cmd.NewState()

	env, err := env.Load(state)
	if err != nil {
		return env.State, err
	}

	if err := boot.Run(env); err != nil {
		return env.State, err
	}

	cmd, err := layout.Unwrap()
	if err != nil {
		return env.State, err
	}

	return env.State, cmd.Fn(env, cmd.Args)
}

// Unwrap unwraps layers to final cmd.
func (layout Layout) Unwrap() (*Cmd, error) {
	args := flag.Args()

	if len(args) == 0 {
		layout.Print()
		return nil, ErrArgMissing
	}

	cmd := layout[args[0]]

	switch {
	case cmd == nil:
		return cmd, fmt.Errorf("%w, %v", ErrArgInvalid, args[0])

	case len(args) <= cmd.Min:
		return nil, cmd.Help

	default:
		cmd.Args = args[1:]
	}

	for pos, arg := range cmd.Args {
		next, ok := cmd.Layout[arg]
		if !ok {
			cmd.Pos = pos
			cmd.Args = args[cmd.Pos+1:]
			break
		}

		cmd = next
	}

	if cmd.Layout != nil || len(cmd.Args) < cmd.Min {
		return nil, cmd.Help
	}

	return cmd, nil
}

// Print prints logo and helps.
func (layout Layout) Print() error {
	fmt.Print(help.Logo)

	for arg, cmd := range layout {
		fmt.Print("\n", arg, "\n")
		fmt.Print(cmd.Help, "\n")
	}

	return nil
}
