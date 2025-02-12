// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/client/git"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// update represents an update.
type update struct {
	srcs    map[string][]string
	secrets *secret.Cfg
	paths   *path.Sigs
}

// Report represents a sig update report.
type Report struct {
	Update interface{}
}

// Update starts a sig update.
func Update(env *env.Env) error {
	if len(env.Cfg.Secrets.Git) == 0 {
		return ErrNoRepos
	}

	updater := &update{
		secrets: env.Cfg.Secrets,
		paths:   env.Paths.Sigs,
	}

	tasks := []func(*secret.Repo) error{
		updater.clone,
		updater.install,
	}

	for _, repo := range updater.secrets.Git {
		for _, task := range tasks {
			if err := task(repo); err != nil {
				return err
			}
		}
	}

	return Refresh(env)
}

// clone clones repo.
func (update *update) clone(repo *secret.Repo) error {
	var (
		owner, _ = filepath.Base(filepath.Dir(repo.URL)), filepath.Base(repo.URL)
		dst      = filepath.Join(update.paths.Tmp, owner)
	)

	switch {
	case dst == "":
		return ErrNoSigTmpDir

	case !filepath.IsAbs(dst):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, dst)
	}

	dot := filepath.Join(dst, ".git")

	// No error for not exist.
	if err := os.RemoveAll(dot); err != nil {
		return fmt.Errorf("%w, %v", fsys.ErrFileDel, dot)
	}

	tag, err := git.Clone(repo, dst)
	if err != nil {
		return err
	}

	slog.Info("clone complete", "tag", tag)

	return update.walk(update.paths.Tmp)
}

// install installs downloaded repo yr src files.
func (update *update) install(_ *secret.Repo) error {
	for dir, files := range update.srcs {
		parent := filepath.Join(update.paths.Src, dir)

		if err := os.MkdirAll(parent, 0750); err != nil {
			return fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, parent)
		}

		for _, file := range files {
			path := filepath.Join(update.paths.Tmp, file)

			stat, err := os.Lstat(path)
			if err != nil {
				return fmt.Errorf("%w, %v", fsys.ErrStat, path)
			}

			// symlinks not allowed.
			if stat.Mode()&os.ModeSymlink != 0 {
				slog.Info(ErrIsSym.Error(), "path", path)
				continue
			}

			dst := filepath.Join(update.paths.Src, dir, filepath.Base(file))

			if err := os.Rename(path, dst); err != nil {
				return fmt.Errorf("%w, %v, %v, %v", fsys.ErrFileCopy, err, file, dst)
			}
		}
	}

	return nil
}

// walk walks a given path to locate yr src files.
func (update *update) walk(path string) error {
	reduced := map[string][]string{}

	srcs, err := fsys.WalkByExt(path, ".yara", ".yar", ".yr")
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrWalk, err, path)
	}

	for _, src := range srcs {
		var (
			rel  = strings.TrimPrefix(src, path)
			base = filepath.Dir(rel)
		)

		reduced[base] = append(reduced[base], rel)
	}

	update.srcs = reduced

	return nil
}

// Mock mocks sigs.
func Mock(env *env.Env) error {
	var (
		rule = `rule eicar : tag { strings: $s1 = "$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!" condition: $s1 }`
		dir  = filepath.Join(env.Paths.Sigs.Src, "mock")
		file = filepath.Join(dir, "mock.yr")
	)

	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	if err := os.WriteFile(file, []byte(rule), 0640); err != nil {
		return err
	}

	return Refresh(env)
}
