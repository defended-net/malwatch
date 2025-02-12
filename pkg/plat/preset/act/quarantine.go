// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Quarantiner represents quarantines.
type Quarantiner struct {
	verb string
	dir  string
}

// NewQuarantiner returns quarantiner for given env.
func NewQuarantiner(env *env.Env) *Quarantiner {
	return &Quarantiner{
		verb: VerbQuarantine,
		dir:  env.Cfg.Acts.Quarantine.Dir,
	}
}

// Load loads a given quarantiner.
func (quarantiner *Quarantiner) Load() error {
	if quarantiner.dir == "" {
		return ErrDisabled
	}

	return nil
}

// Act quarantines hits for given result.
func (quarantiner *Quarantiner) Act(result *state.Result) error {
	if quarantiner.dir == "" {
		return ErrQuarantineNoDir
	}

	for path, meta := range result.Paths {
		dst := fsys.QuarantinePath(quarantiner.dir, path)

		attr := &fsys.Attr{
			UID:  meta.Attr.UID,
			GID:  meta.Attr.GID,
			Mode: meta.Attr.Mode,
		}

		meta.Status = filepath.Base(dst)

		if err := fsys.Mv(path, dst, attr); err != nil {
			result.Errs.Add(err)
		}
	}

	return nil
}

// Verb returns a given quarantiner verb.
func (quarantiner *Quarantiner) Verb() string {
	return quarantiner.verb
}
