// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// IsExp verifies if given file has exceeded mtime. Timestomp protected.
func IsExp(expiry time.Time, fd int) (bool, *unix.Stat_t) {
	stat := &unix.Stat_t{}

	if err := unix.Fstat(fd, stat); err != nil {
		return false, nil
	}

	var (
		cTime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
		mTime = time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
	)

	if cTime.Before(expiry) && mTime.Before(expiry) {
		return true, nil
	}

	return false, stat
}

// IsRel verifies if given path is rel to given base paths.
func IsRel(path string, bases ...string) bool {
	for _, base := range bases {
		rel, err := filepath.Rel(base, path)
		if err == nil && filepath.IsLocal(rel) {
			return true
		}
	}

	return false
}

// HasDotDots validates given paths for dot dots, rel, root or curr dir.
func HasDotDots(paths ...string) error {
	for _, path := range paths {
		// Check segments before collapse.
		if slices.Contains(strings.Split(filepath.ToSlash(path), "/"), "..") {
			return fmt.Errorf("%w, %v", ErrPathTraverse, path)
		}

		clean := filepath.Clean(path)

		if !filepath.IsAbs(clean) {
			return fmt.Errorf("%w, %v", ErrPathNotAbs, path)
		}

		if clean == "/" {
			return fmt.Errorf("%w, %v", ErrPathRoot, path)
		}
	}

	return nil
}

// RootName name of given path rel to given root.
func RootName(root *os.Root, path string) (string, error) {
	if root == nil {
		return "", ErrPathRoot
	}

	name := path

	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(root.Name(), path)
		if err != nil {
			return "", err
		}

		name = rel
	}

	name = filepath.Clean(name)

	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("%w, %v", ErrPathLocal, path)
	}

	return name, nil
}
