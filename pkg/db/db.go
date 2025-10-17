// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Load loads the db.
func Load(env *env.Env) error {
	switch {
	// Disabled, not an error.
	case env.Cfg.Database.Dir == "":
		return nil

	case !filepath.IsAbs(env.Cfg.Database.Dir):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, env.Cfg.Database.Dir)
	}

	if err := os.MkdirAll(env.Cfg.Database.Dir, 0700); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, env.Cfg.Database.Dir)
	}

	db, err := bbolt.Open(env.Paths.Install.Db, 0600, nil)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrOpen, err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("hits")); err != nil {
			return fmt.Errorf("%w, %v", ErrBktCreate, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%w, %v", ErrTxUpdate, err)
	}

	env.Db = db

	return nil
}
