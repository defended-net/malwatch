// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package git

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

type commit struct {
	commit *object.Commit
	tag    string
}

// Clone clones a repo. The clone's tag is used for log reference.
func Clone(repo *secret.Repo, dst string) (string, error) {
	opts := &git.CloneOptions{
		URL:          repo.URL,
		Depth:        1,
		SingleBranch: true,

		Auth: &http.BasicAuth{
			Username: repo.User,
			Password: repo.Token,
		},
	}

	clone, err := git.PlainClone(dst, false, opts)
	if err != nil {
		return "", fmt.Errorf("%w, %v, %v", ErrClone, err, repo.URL)
	}

	tag, err := latestTag(clone)
	if err != nil {
		return "", fmt.Errorf("%w, %v", ErrRefTag, err)
	}

	return tag, nil
}

// latestTag returns most recent release.
func latestTag(repo *git.Repository) (string, error) {
	latest := &commit{
		commit: &object.Commit{},
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("%w, %v", ErrRefTag, err)
	}

	if err := tags.ForEach(func(ref *plumbing.Reference) error {
		rev := plumbing.Revision(ref.Name().String())

		hash, err := repo.ResolveRevision(rev)
		if err != nil {
			return err
		}

		info, err := repo.CommitObject(*hash)
		if err != nil {
			return err
		}

		if info.Committer.When.After(latest.commit.Committer.When) {
			latest.commit = info
			latest.tag = ref.Name().String()
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("%w, %v", ErrRefIter, err)
	}

	return filepath.Base(latest.tag), nil
}
