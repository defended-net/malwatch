// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/client/s3"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Exiler represents an acter for exiles.
type Exiler struct {
	verb          string
	secrets       *secret.S3
	transport     *s3.Transport
	quarantineDir string
}

// NewExiler returns exiler for given env.
func NewExiler(env *env.Env) *Exiler {
	return &Exiler{
		verb:          VerbExile,
		secrets:       env.Cfg.Secrets.S3,
		quarantineDir: env.Cfg.Acts.Quarantine.Dir,
	}
}

// Load loads a given exiler.
func (exiler *Exiler) Load(_ *os.Root) error {
	if exiler.secrets.Endpoint == "" {
		return acter.ErrDisabled
	}

	transport, err := s3.New(exiler.secrets)
	if err != nil {
		return err
	}

	exiler.transport = transport

	return nil
}

// Act exiles hits from given result.
func (exiler *Exiler) Act(result *state.Result) error {
	if exiler.secrets.Region == "" {
		return ErrExileNoRegion
	}

	for path, meta := range result.Paths {
		exiler.Single(result, path, meta)
	}

	return nil
}

// Single exiles single path.
func (exiler *Exiler) Single(result *state.Result, path string, meta *hit.Meta) {
	path = filepath.Clean(path)

	if meta.Attr == nil {
		result.AddErr(fmt.Errorf("%w, %v", ErrAttrInvalid, path))

		return
	}

	fd, _, err := fsys.Open(path)
	if err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrExileDelErr, err, path))

		return
	}

	file := os.NewFile(uintptr(fd), path)
	defer fsys.Close(file)

	if err := exiler.transport.Ul(path, file); err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrExileUpload, err, path))

		return
	}

	meta.Status = s3.Scheme + filepath.Base(path)

	switch {
	case slices.Contains(meta.Acts, VerbQuarantine) ||
		slices.Contains(meta.Acts, VerbClean):

		return

	case fsys.Unlink(path, false) != nil:
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrExileDelErr, err, path))

		return
	}

	slog.Info("deleted", "path", path)
}

// Verb returns a given exiler verb.
func (exiler *Exiler) Verb() string {
	return exiler.verb
}
