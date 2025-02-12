// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package history

import (
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/tui/tbl"
)

// Get prints hit history for given target or path.
// ./malwatch history get [TARGET | PATH]
func Get(env *env.Env, args []string) error {
	switch {
	case len(args) == 0:
		histories, err := hit.SelectAll(env.Db)
		if err != nil {
			return err
		}

		if len(histories) == 0 {
			return tbl.Print("", tbl.HdrAppHit, nil)
		}

		for _, history := range histories {
			tbl.Print(history.Target, tbl.HdrAppHit, history.Paths.ToSlice())
		}

	case filepath.IsAbs(args[0]):
		var (
			path   = args[0]
			target = re.Target(args[0])
			cells  = [][]string{}
		)

		meta, err := hit.SelectLast(env.Db, path)
		if err != nil {
			return err
		}

		if row := meta.ToSlice(path); len(row) > 0 {
			cells = [][]string{meta.ToSlice(path)}
		}

		return tbl.Print(target, tbl.HdrAppHit, cells)

	case len(args) == 1:
		path := args[0]

		paths, err := hit.SelectTarget(env.Db, path)
		if err != nil {
			return err
		}

		return tbl.Print(path, tbl.HdrAppHit, paths.ToSlice())
	}

	return nil
}

// Del deletes hit history for given target or path.
// ./malwatch history del [TARGET | PATH]
func Del(env *env.Env, args []string) error {
	input := args[0]

	if filepath.IsAbs(input) {
		target := re.Target(input)

		return hit.DelPath(env.Db, target, input)
	}

	return hit.DelTarget(env.Db, input)
}
