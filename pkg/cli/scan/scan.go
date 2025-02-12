// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import (
	"fmt"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/scan"
)

// Do starts a scan.
func Do(env *env.Env, args []string) error {
	switch {
	case len(args) == 0:
		paths, err := scan.Glob(env.Cfg.Scans.Paths)
		if err != nil {
			return err
		}

		args = paths

	case !filepath.IsAbs(args[0]):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, args[0])
	}

	scan, err := scan.New(env, args...)
	if err != nil {
		return err
	}

	return scan.Run()
}
