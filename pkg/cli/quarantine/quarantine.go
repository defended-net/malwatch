// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package quarantine

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Do quarantines.
// ./malwatch quarantine [PATH]
func Do(env *env.Env, args []string) error {
	var (
		path = args[0]
		stat = &unix.Stat_t{}
	)

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, args[0])
	}

	if err := unix.Stat(path, stat); err != nil {
		return err
	}

	result := &state.Result{
		Target: re.Target(path),
		Paths: state.Paths{
			path: hit.NewMeta(fsys.NewAttr(stat), []string{}, "quarantine"),
		},
	}

	acter, err := acter.Get(env.Plat.Acters(), "quarantine")
	if err != nil {
		return err
	}

	if err := acter.Act(result); err != nil {
		return err
	}

	return result.Save(env.Db)
}
