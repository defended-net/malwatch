// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exile

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Do exiles.
// ./malwatch exile [PATH]
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

	var (
		paths = state.Paths{
			path: hit.NewMeta(fsys.NewAttr(stat), []string{}, "exile"),
		}

		result = state.NewResult(re.Target(path), paths)
	)

	if err := acter.Do(env.Plat.Acters(), "exile", result); err != nil {
		return err
	}

	for _, err := range result.Errs() {
		slog.Error(err.Error())
	}

	return result.Save(env.Db)
}
