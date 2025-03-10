// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package restore

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/client/s3"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Do restores.
// ./malwatch restore [PATH]
func Do(env *env.Env, args []string) error {
	path := args[0]

	if err := fsys.HasDotDots(path); err != nil {
		return err
	}

	hit, err := hit.SelectLast(env.Db, path)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(hit.Status, s3.Scheme):
		transport, err := s3.New(env.Cfg.Secrets.S3)
		if err != nil {
			return err
		}

		return transport.Dl(path, hit.Attr)

	case filepath.IsAbs(hit.Status):
		return hit.Restore(env.Cfg.Acts.Quarantine.Dir, path)

	default:
		return fmt.Errorf("%w, %v", ErrStatusUnknown, hit.Status)
	}
}
