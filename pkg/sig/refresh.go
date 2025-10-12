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

// writeIdx writes the yr index for given update.
func (update *update) writeIdx() error {
	slog.Info("writing yara index", "path", update.paths.Idx)

	file, err := os.Create(update.paths.Idx)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileCreate, err, update.paths.Idx)
	}
	defer file.Close()

	wr := bufio.NewWriter(file)

	for _, paths := range update.srcs {
		for _, path := range paths {
			if _, err := wr.WriteString(`include "` + filepath.Join(update.paths.Src, path) + "\"\n"); err != nil {
				return fmt.Errorf("%w, %v, %v", ErrYrIdxWrite, err, update.paths.Idx)
			}
		}
	}

	return wr.Flush()
}

// compile saves the given update's index as bytecode.
func (update *update) compile() error {
	if err := update.writeIdx(); err != nil {
		return err
	}

	idx, err := os.Open(update.paths.Idx)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, update.paths.Idx)
	}
	defer idx.Close()

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
		return fmt.Errorf("%w, %v", ErrYrRulesGet, err)
	}

	slog.Info("saving rules", "path", update.paths.Yrc)

	if err = rules.Save(update.paths.Yrc); err != nil {
		return fmt.Errorf("%w, %v", ErrYrRulesSave, err)
	}

	if err := os.Chmod(update.paths.Yrc, 0600); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, update.paths.Yrc)
	}

	return nil
}
