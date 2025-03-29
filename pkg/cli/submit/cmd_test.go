// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package submit

import (
	"fmt"
	"testing"

	"github.com/defended-net/malwatch/pkg/cli"
	"github.com/defended-net/malwatch/pkg/cli/help"
)

func TestCmd(t *testing.T) {
	var (
		want = &cli.Cmd{
			Help: help.ErrSubmit,
			Min:  1,
			Fn:   Do,
		}

		got = Cmd()
	)

	// Args field is []string, will have to iteratively compare.

	if got.Help != want.Help {
		t.Errorf("unexpected cmd help result %v, want %v", got.Help, want.Help)
	}

	if got.Min != want.Min {
		t.Errorf("unexpected cmd min result %v, want %v", got.Min, want.Min)
	}

	if fmt.Sprintf("%v", got.Fn) != fmt.Sprintf("%v", want.Fn) {
		t.Errorf("unexpected cmd fn result %v, want %v", got.Fn, want.Fn)
	}
}
