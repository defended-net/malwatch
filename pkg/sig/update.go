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
	Update any
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

// delDotGit del .git through update.paths.Root after proving target is abs, has no traversal,
// is under sig tmp dir and is local to install update.paths.Root.
func (update *update) delDotGit(dst string) error {
	dot := filepath.Join(dst, ".git")

	if err := fsys.HasDotDots(dot); err != nil {
		return err
	}

	if !fsys.IsRel(dot, update.paths.Tmp) {
		return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, dot)
	}

	rel, err := fsys.RootName(update.paths.Root, dot)
	if err != nil {
		return err
	}

	if err := update.paths.Root.RemoveAll(rel); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileDel, err, dot)
	}

	return nil
}

// clone clones repo.
func (update *update) clone(repo *secret.Repo) error {
	var (
		owner = filepath.Base(filepath.Dir(repo.URL))
		dst   = filepath.Join(update.paths.Tmp, owner)
	)

	if owner == "" || owner == "." || owner == "/" {
		return ErrNoRepoOwner
	}

	if err := fsys.HasDotDots(dst); err != nil {
		return err
	}

	switch {
	case dst == "":
		return ErrNoSigTmpDir

	case !filepath.IsAbs(dst):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, dst)

	case !fsys.IsRel(dst, update.paths.Tmp):
		return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, dst)
	}

	if err := update.delDotGit(dst); err != nil {
		return err
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
	if update.paths.Root == nil {
		return fsys.ErrPathRoot
	}

	attr := &fsys.Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	for dir, files := range update.srcs {
		parent := filepath.Join(update.paths.Src, dir)

		if err := fsys.HasDotDots(parent); err != nil {
			return err
		}

		// Verify install dir is within expected src dir.
		if !fsys.IsRel(parent, update.paths.Src) {
			return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, parent)
		}

		parentName, err := fsys.RootName(update.paths.Root, parent)
		if err != nil {
			return err
		}

		if err := update.paths.Root.MkdirAll(parentName, 0700); err != nil {
			return fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, parent)
		}

		for _, file := range files {
			var (
				src = filepath.Join(update.paths.Tmp, file)
				dst = filepath.Join(update.paths.Src, dir, filepath.Base(file))
			)

			if err := fsys.HasDotDots(src, dst); err != nil {
				return err
			}

			switch {
			case !fsys.IsRel(src, update.paths.Tmp):
				return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, src)

			case !fsys.IsRel(dst, update.paths.Src):
				return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, dst)
			}

			if err := fsys.MvToRoot(update.paths.Root, src, dst, attr); err != nil {
				return fmt.Errorf("%w, %v, %v", fsys.ErrFileCopy, err, dst)
			}
		}
	}

	return nil
}

// walk walks a given path to locate yr src files.
func (update *update) walk(path string) error {
	reduced := map[string][]string{}

	path = strings.TrimRight(filepath.Clean(path), "/") + "/"

	srcs, err := fsys.WalkByExt(path, ".yara", ".yar", ".yr")
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrWalk, err, path)
	}

	for _, src := range srcs {
		var (
			parent = strings.TrimPrefix(src, path)
			dir    = filepath.Dir(parent)
		)

		if strings.Contains(parent, "..") {
			slog.Info(fsys.ErrPathTraverse.Error(), "path", src)
			continue
		}

		reduced[dir] = append(reduced[dir], parent)
	}

	update.srcs = reduced

	return nil
}

// Mock mocks sigs.
func Mock(env *env.Env, monitor bool) error {
	var (
		rule = `rule eicar : tag { strings: $s1 = "$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!" condition: $s1 }`
		dir  = filepath.Join(env.Paths.Sigs.Src, "mock")
		file = filepath.Join(dir, "mock.yr")
	)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := os.WriteFile(file, []byte(rule), 0600); err != nil {
		return err
	}

	if err := Refresh(env); err != nil {
		return err
	}

	if monitor {
		return Monitor(env)
	}

	return nil
}
