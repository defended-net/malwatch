// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Quarantiner represents quarantines.
type Quarantiner struct {
	verb string
	dir  string
	root *os.Root
}

// NewQuarantiner returns quarantiner for given env.
func NewQuarantiner(env *env.Env) *Quarantiner {
	return &Quarantiner{
		verb: VerbQuarantine,
		dir:  env.Cfg.Acts.Quarantine.Dir,
	}
}

// Load loads a given quarantiner.
func (quarantiner *Quarantiner) Load(_ *os.Root) error {
	if quarantiner.dir == "" {
		return acter.ErrDisabled
	}

	root, err := os.OpenRoot(quarantiner.dir)
	if err != nil {
		return err
	}

	quarantiner.root = root

	return nil
}

// Act quarantines hits for given result.
func (quarantiner *Quarantiner) Act(result *state.Result) error {
	if quarantiner.dir == "" {
		return ErrQuarantineNoDir
	}

	if quarantiner.root == nil {
		return fsys.ErrPathRoot
	}

	for path, meta := range result.Paths {
		dst := fsys.QuarantinePath(quarantiner.dir, path)

		meta.Status = filepath.Base(dst)

		if err := fsys.MvToRoot(quarantiner.root, path, dst, meta.Attr); err != nil {
			// try next one.
			result.AddErr(fmt.Errorf("%w, %v", ErrQuarantineMv, err))
		}
	}

	return nil
}

// Verb returns a given quarantiner verb.
func (quarantiner *Quarantiner) Verb() string {
	return quarantiner.verb
}
