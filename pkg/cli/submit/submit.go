// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package submit

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/client/http"
	"github.com/defended-net/malwatch/pkg/fsys"
	"golang.org/x/sys/unix"
)

// Do submits.
// ./malwatch submit [PATH]
func Do(env *env.Env, args []string) error {
	var (
		path = args[0]
		stat = &unix.Stat_t{}
	)

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, path)
	}

	if err := unix.Stat(path, stat); err != nil {
		return err
	}

	if err := http.Submit(env.Cfg.Secrets.Submit, path); err != nil {
		return err
	}

	slog.Info("submit: success")

	return nil
}
