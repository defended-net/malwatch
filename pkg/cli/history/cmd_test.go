// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package history

import (
	"fmt"
	"testing"

	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

func TestCmd(t *testing.T) {
	var (
		want = &cli.Cmd{
			Help: help.ErrHistory,
			Min:  1,

			Sub: cli.Sub{
				"get": {
					Help: help.ErrHistory,
					Fn:   Get,
				},

				"del": {
					Help: help.ErrHistory,
					Min:  1,
					Fn:   Del,
				},
			},
		}

		got = Cmd()

		input = []string{
			"get",
			"del",
		}
	)

	// Args field is []string, will have to iteratively compare.

	if got.Help != want.Help {
		t.Errorf("unexpected cmd help result %v, want %v", got.Help, want.Help)
	}

	if got.Min != want.Min {
		t.Errorf("unexpected cmd min result %v, want %v", got.Min, want.Min)
	}

	for _, input := range input {
		if fmt.Sprintf("%v", got.Sub[input].Fn) != fmt.Sprintf("%v", want.Sub[input].Fn) {
			t.Errorf("unexpected cmd route fn result %v, want %v", got.Sub[input].Fn, want.Sub[input].Fn)
		}
	}
}
