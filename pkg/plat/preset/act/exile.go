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
func (exiler *Exiler) Load() error {
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

// Act exiles hits from a given result.
func (exiler *Exiler) Act(result *state.Result) error {
	if exiler.secrets.Region == "" {
		return ErrExileNoRegion
	}

	for path, meta := range result.Paths {
		if err := exiler.transport.Ul(path); err != nil {
			// try next one.
			result.AddErr(fmt.Errorf("%w, %v, %v", ErrExileUpload, err, path))
			continue
		}

		meta.Status = s3.Scheme + filepath.Base(path)

		if slices.Contains(meta.Acts, VerbQuarantine) || slices.Contains(meta.Acts, VerbClean) {
			continue
		}

		if err := os.Remove(path); err != nil {
			// try next one.
			result.AddErr(fmt.Errorf("%w, %v, %v", ErrExileDelErr, err, path))
			continue
		}

		slog.Info("deleted", "path", path)
	}

	return nil
}

// Verb returns a given exiler verb.
func (exiler *Exiler) Verb() string {
	return exiler.verb
}
