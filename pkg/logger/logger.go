// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

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

	if path == "" {
		return nil
	}

	dirName, err := fsys.RootName(env.Paths.Install.Root, dir)
	if err != nil {
		return err
	}

	if err := env.Paths.Install.Root.MkdirAll(dirName, 0700); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, dir)
	}

	name, err := fsys.RootName(env.Paths.Install.Root, path)
	if err != nil {
		return err
	}

	file, err := env.Paths.Install.Root.OpenFile(name, unix.O_RDWR|unix.O_CREAT|unix.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrOpen, path)
	}

	writers = append(writers, file)

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
