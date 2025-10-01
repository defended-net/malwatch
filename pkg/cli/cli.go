// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"flag"
	"fmt"
	"sort"

	"github.com/defended-net/malwatch/pkg/boot"
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/cli/help"
	"github.com/defended-net/malwatch/pkg/cmd"
)

// Cmd represents a cmd.
type Cmd struct {
	Help error
	Min  int
	Sub  Sub
	Fn   fn
	args []string
	pos  int
}

// Sub represents subcmd layout.
type Sub map[string]*Cmd

// fn represents a cmdfn.
type fn func(*env.Env, []string) error

// Run boots and then starts the cmdfn from given subcmd.
func Run(sub Sub) (*cmd.State, error) {
	state := cmd.NewState()

	env, err := env.Load(state)
	if err != nil {
		return state, err
	}

	if err := boot.Run(env); err != nil {
		return env.State, err
	}

	cmd, err := sub.Route()
	if err != nil {
		return env.State, err
	}

	return env.State, cmd.Fn(env, cmd.args)
}

// Route returns follow up cmd for given subcmd.
func (sub Sub) Route() (*Cmd, error) {
	args := flag.Args()

	cmd, err := base(sub, args)
	if err != nil {
		return nil, err
	}

	for pos, arg := range cmd.args {
		next, ok := cmd.Sub[arg]
		if !ok {
			cmd.pos = pos
			cmd.args = args[cmd.pos+1:]
			break
		}

		cmd = next
	}

	if cmd.Sub != nil || len(cmd.args) < cmd.Min {
		return nil, cmd.Help
	}

	return cmd, nil
}

// base returns the entry cmd.
func base(sub Sub, args []string) (*Cmd, error) {
	if len(args) == 0 {
		sub.Print()
		return nil, ErrArgNone
	}

	cmd := sub[args[0]]
	if cmd == nil {
		return nil, fmt.Errorf("%w, %v", ErrArgInvalid, args[0])
	}

	if len(args) <= cmd.Min {
		return nil, cmd.Help
	}

	cmd.args = args[1:]

	return cmd, nil
}

// Print prints logo and helps.
func (sub Sub) Print() error {
	fmt.Print(help.Logo)

	args := make([]string, 0, len(sub))
	for arg := range sub {
		args = append(args, arg)
	}

	sort.Strings(args)

	for _, arg := range args {
		cmd := sub[arg]
		fmt.Print("\n", arg, "\n")
		fmt.Print(cmd.Help, "\n")
	}

	return nil
}
