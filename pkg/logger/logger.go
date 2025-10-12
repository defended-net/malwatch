// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Load loads the logger.
func Load(env *env.Env) error {
	var (
		path = env.Paths.Install.Log
		dir  = filepath.Dir(path)

		writers = []io.Writer{}

		opts = &slog.HandlerOptions{
			AddSource:   env.Cfg.Log.Verbose,
			ReplaceAttr: rewriteAttrs,
		}

		attrs = []slog.Attr{
			slog.String("host", env.Cfg.Identifier),
		}
	)

	if filepath.IsAbs(path) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, dir)
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("%w, %v, %v", ErrOpen, err, path)
		}

		writers = append(writers, file)
	}

	if !env.Opts.Unattended {
		writers = append(writers, os.Stdout)
	}

	var (
		writer  = io.MultiWriter(writers...)
		handler = slog.NewJSONHandler(writer, opts).WithAttrs(attrs)
	)

	slog.SetDefault(slog.New(handler))

	return nil
}

// rewriteAttrs rewrites attrs.
func rewriteAttrs(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.SourceKey {
		attr.Value = slog.StringValue(filepath.Base(attr.Value.String()))
	}

	return attr
}
