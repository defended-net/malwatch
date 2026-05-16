// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/third_party/yr"
)

// Refresh starts a sig refresh.
func Refresh(env *env.Env) error {
	updater := &update{
		secrets: env.Cfg.Secrets,
		paths:   env.Paths.Sigs,
	}

	tasks := []func(*update) error{
		func(_ *update) error { return updater.walk(updater.paths.Src) },
		func(_ *update) error { return updater.compile() },
	}

	for _, task := range tasks {
		if err := task(updater); err != nil {
			return err
		}
	}

	return nil
}

// writeIdx writes yr index for given update through update.paths.Root and returns
// the open file.
func (update *update) writeIdx() (*os.File, error) {
	slog.Info("writing yara index", "path", update.paths.Idx)

	name, err := fsys.RootName(update.paths.Root, update.paths.Idx)
	if err != nil {
		return nil, fmt.Errorf("%w, %v, %v", fsys.ErrFileCreate, err, update.paths.Idx)
	}

	file, err := update.paths.Root.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("%w, %v, %v", fsys.ErrFileCreate, err, update.paths.Idx)
	}

	var (
		keep bool
		wr   = bufio.NewWriter(file)
	)

	defer func() {
		if !keep {
			fsys.Close(file)
		}
	}()

	for _, paths := range update.srcs {
		for _, path := range paths {
			if _, err := wr.WriteString(`include "` + filepath.Join(update.paths.Src, path) + "\"\n"); err != nil {
				return nil, fmt.Errorf("%w, %v, %v", ErrYrIdxWrite, err, update.paths.Idx)
			}
		}
	}

	if err := wr.Flush(); err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, update.paths.Idx)
	}

	keep = true

	return file, nil
}

// compile saves given index as bytecode.
func (update *update) compile() error {
	idx, err := update.writeIdx()
	if err != nil {
		return err
	}
	defer fsys.Close(idx)

	yr, err := yr.NewCompiler()
	if err != nil {
		return fmt.Errorf("%w, %v", ErrYrCompiler, err)
	}

	if err := yr.AddFile(idx, "index"); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrYrIdxAdd, err, update.paths.Idx)
	}

	slog.Info("compiling rules", "path", update.paths.Idx)

	rules, err := yr.GetRules()
	if err != nil {
		return fmt.Errorf("%w, %v", ErrYrcGet, err)
	}

	slog.Info("saving rules", "path", update.paths.Yrc)

	if err = rules.Save(update.paths.Yrc); err != nil {
		return fmt.Errorf("%w, %v", ErrYrcSave, err)
	}

	yrcName, err := fsys.RootName(update.paths.Root, update.paths.Yrc)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, update.paths.Yrc)
	}

	if err := update.paths.Root.Chmod(yrcName, 0600); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, update.paths.Yrc)
	}

	return nil
}
