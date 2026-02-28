// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// Attr represents file attributes.
type Attr struct {
	UID   int         `json:"uid"`
	GID   int         `json:"gid"`
	Mode  fs.FileMode `json:"mode"`
	CTime time.Time   `json:"ctime"`
	MTime time.Time   `json:"mtime"`
}

// NewAttr returns an attr from given stat.
func NewAttr(stat *unix.Stat_t) *Attr {
	return &Attr{
		UID:   int(stat.Uid),
		GID:   int(stat.Gid),
		Mode:  fs.FileMode(stat.Mode).Perm(),
		CTime: time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC(),
		MTime: time.Unix(stat.Mtim.Sec, stat.Ctim.Nsec).UTC(),
	}
}

// Mv moves files across different mnts. uid and gid only apply for files.
func Mv(srcPath string, dstPath string, attr *Attr) error {
	var (
		src  = filepath.Clean(srcPath)
		dst  = filepath.Clean(dstPath)
		stat = &unix.Stat_t{}
	)

	if attr == nil {
		return fmt.Errorf("%w, %v", ErrAttrInvalid, src)
	}

	if err := HasDotDots(src, dst); err != nil {
		return err
	}

	if err := unix.Lstat(src, stat); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrStat, err, src)
	}

	switch stat.Mode & unix.S_IFMT {
	// Regular file, proceed.
	case unix.S_IFREG:

	// Reject symlink.
	case unix.S_IFLNK:
		return fmt.Errorf("%w, %v", ErrIsSym, src)

	// Reject dir.
	case unix.S_IFDIR:
		return fmt.Errorf("%w, %v", ErrIsDir, src)

	default:
		return fmt.Errorf("%w, %v", ErrIsNotReg, src)
	}

	if err := func() error {
		srcF, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, src)
		}
		defer srcF.Close()

		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			return fmt.Errorf("%w, %v, %v", ErrDirCreate, err, dst)
		}

		dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, attr.Mode)
		if err != nil {
			return fmt.Errorf("%w, %v, %v", ErrFileCreate, err, dst)
		}
		defer dstF.Close()

		if _, err = io.Copy(dstF, srcF); err != nil {
			return fmt.Errorf("%w, %v, %v, %v", ErrFileCopy, err, src, dst)
		}

		if err := os.Chown(dst, attr.UID, attr.GID); err != nil {
			// Might be unpriv user, which is fine.s
			slog.Info(ErrChown.Error(), "path", dst)
		}

		if err = os.Chmod(dst, attr.Mode); err != nil {
			slog.Info(ErrChmod.Error(), "path", dst)
		}

		if err := dstF.Sync(); err != nil {
			return fmt.Errorf("%w, %v", ErrFileSync, dst)
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileDel, err, src)
	}

	slog.Info("moved", "src", src, "dst", dst)

	return nil
}

// Walk returns a recursive file list for a path.
func Walk(path string) ([]string, error) {
	paths := []string{}

	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return fmt.Errorf("%w, %v, %v", ErrWalk, err, path)

		case entry.IsDir():
			return nil
		}

		paths = append(paths, path)

		return nil
	})

	return paths, err
}

// WalkByExt returns a recursive file list for a given path and extesion(s).
func WalkByExt(path string, exts ...string) ([]string, error) {
	result := []string{}

	paths, err := Walk(path)
	if err != nil {
		return result, fmt.Errorf("%w, %v, %v", ErrWalk, err, path)
	}

	for _, path := range paths {
		if slices.Contains(exts, filepath.Ext(path)) {
			result = append(result, path)
		}
	}

	return result, nil
}

// QuarantinePath returns a path's quarantine path from given parent dir and detection path.
func QuarantinePath(quarantineDir string, path string) string {
	var (
		dir, file = filepath.Split(path)
		renamed   = fmt.Sprintf("%s-%d", file, time.Now().Unix())
	)

	return filepath.Join(quarantineDir, dir, renamed)
}

// MntPoint returns a given path's mnt point.
func MntPoint(path string) (string, error) {
	cur := &unix.Stat_t{}

	if err := unix.Stat(path, cur); err != nil {
		return "", fmt.Errorf("%w, %v, %v", ErrStat, err, path)
	}

	for path != "/" {
		var (
			parent = filepath.Dir(path)
			stat   = &unix.Stat_t{}
		)

		if err := unix.Stat(parent, stat); err != nil {
			return "", fmt.Errorf("%w, %v, %v", ErrStat, err, parent)
		}

		if stat.Dev != cur.Dev {
			return path, nil
		}

		path = parent
		cur = stat
	}

	return "/", nil
}

// IsRel verifies if a path is relative to a base path.
func IsRel(path string, bases ...string) bool {
	for _, base := range bases {
		rel, err := filepath.Rel(base, path)

		if err == nil && !strings.HasPrefix(rel, "../") && rel != ".." {
			return true
		}
	}

	return false
}

// HasDotDots validates paths against dot dots, relative, root or current dir.
func HasDotDots(paths ...string) error {
	for _, path := range paths {
		// Check segments before collapse.
		if slices.Contains(strings.Split(filepath.ToSlash(path), "/"), "..") {
			return fmt.Errorf("%w, %v", ErrPathTravers, path)
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

// IsExp verifies is a file has exceeded max mtime. Timestomp protection with ctime.
func IsExp(expiry time.Time, fd int) (bool, *unix.Stat_t) {
	stat := &unix.Stat_t{}

	if err := unix.Fstat(fd, stat); err != nil {
		return false, nil
	}

	var (
		cTime = time.Unix(stat.Ctim.Sec, 0)
		mTime = time.Unix(stat.Mtim.Sec, 0)
	)

	if cTime.Before(expiry) && mTime.Before(expiry) {
		return true, nil
	}

	return false, stat
}
