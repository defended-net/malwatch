// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	"github.com/BurntSushi/toml"
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
		CTime: time.Unix(stat.Ctim.Sec, 0).UTC(),
		MTime: time.Unix(stat.Mtim.Sec, 0).UTC(),
	}
}

// Mv moves files across different mnts. uid and gid only apply for files.
func Mv(srcPath string, dstPath string, attr *Attr) error {
	var (
		src = filepath.Clean(srcPath)
		dst = filepath.Clean(dstPath)
	)

	if err := HasDotDots(src, dst); err != nil {
		return err
	}

	srcF, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, src)
	}
	defer srcF.Close()

	stat, err := srcF.Stat()
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrStat, err, src)
	}

	if stat.IsDir() {
		return fmt.Errorf("%w, %v", ErrDirMv, src)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrDirCreate, err, dst)
	}

	dstF, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileCreate, err, dst)
	}
	defer dstF.Close()

	// Replicate file.
	if _, err = io.Copy(dstF, srcF); err != nil {
		return fmt.Errorf("%w, %v, %v, %v", ErrFileCopy, err, src, dst)
	}

	if err := os.Chown(dst, attr.UID, attr.GID); err != nil {
		// Do not return, it might be running as ordinary user.
		slog.Info(ErrChown.Error(), "path", dst)
	}

	if err = os.Chmod(dst, attr.Mode); err != nil {
		slog.Info(ErrChmod.Error(), "path", dst)
	}

	// Delete src file.
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
		for _, ext := range exts {
			if filepath.Ext(path) == ext {
				result = append(result, path)
				break
			}
		}
	}

	return result, nil
}

// InstallTOML installs a toml file by first checking if a matching .toml file exists and if not,
// will write the file but with file extension .disabled.
func InstallTOML(path string, cfg any) error {
	// .toml exists, abort.
	if _, err := toml.DecodeFile(path, cfg); err == nil {
		// Should not be logged.
		return fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	disabled := strings.Replace(path, ".toml", ".disabled", 1)

	// .disabled exists, abort.
	if _, err := os.Stat(disabled); err == nil {
		return nil
	}

	if err := WriteTOML(disabled, cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLWrite, err, disabled)
	}

	return nil
}

// ReadTOML reads a toml file for given cfg.
func ReadTOML(path string, cfg any) error {
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	return nil
}

// WriteTOML (over)writes a toml file with given cfg.
func WriteTOML(path string, cfg any) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, path)
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLWrite, err, path)
	}

	return nil
}

// QuarantinePath returns a path's quarantine path from given parent dir and detection path.
func QuarantinePath(quarantineDir string, path string) string {
	srcD, srcF := filepath.Split(path)

	return filepath.Join(quarantineDir, srcD, srcF+"-"+strconv.FormatInt(time.Now().Unix(), 10))
}

// MntPoint returns a given path's mnt point.
func MntPoint(path string) (string, error) {
	stat := &unix.Stat_t{}

	if err := unix.Stat(path, stat); err != nil {
		return "", fmt.Errorf("%w, %v, %v", ErrStat, err, path)
	}

	for {
		if path == "/" {
			return "/", nil
		}

		var (
			parent     = filepath.Dir(path)
			parentStat = &unix.Stat_t{}
		)

		if err := unix.Stat(path, stat); err != nil {
			return "", fmt.Errorf("%w, %v, %v", ErrStat, err, path)
		}

		if parentStat.Dev != stat.Dev {
			break
		}

		path = parent
		stat = parentStat
	}

	return path, nil
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
		path = filepath.Clean(path)

		// filepath.Clean will transform to shortest path, which can still be abused.
		if !filepath.IsAbs(path) {
			return fmt.Errorf("%w, %v", ErrPathNotAbs, path)
		}

		if path == "/" {
			return fmt.Errorf("%w, %v", ErrPathRoot, path)
		}
	}

	return nil
}

// IsExpired verifies is a file has exceeded a max mtime. Includes timestomp protection for ctime.
func IsExpired(expiry time.Time, path string) (*unix.Stat_t, bool) {
	stat := &unix.Stat_t{}

	if err := unix.Stat(path, stat); err != nil {
		return stat, true
	}

	cTime, mTime := time.Unix(stat.Ctim.Sec, 0), time.Unix(stat.Mtim.Sec, 0)

	if cTime.Before(expiry) && mTime.Before(expiry) {
		return stat, true
	}

	return stat, false
}
