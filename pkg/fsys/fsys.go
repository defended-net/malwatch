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
		MTime: time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec).UTC(),
	}
}

// Open opens given path and returns rdonly fd and stat.
func Open(path string) (int, *unix.Stat_t, error) {
	return OpenFile(path, unix.O_RDONLY)
}

// OpenFile opens given path and returns fd and stat.
func OpenFile(path string, flags int) (int, *unix.Stat_t, error) {
	if err := HasDotDots(path); err != nil {
		return -1, nil, err
	}

	var (
		keep   bool
		mode   uint64
		stat   = &unix.Stat_t{}
		clean  = filepath.Clean(path)
		parent = filepath.Dir(clean)
		name   = filepath.Base(clean)
	)

	if flags&unix.O_CREAT != 0 {
		mode = 0600
	}

	parentFd, err := openDir(parent)
	if err != nil {
		return -1, nil, ErrFileOpen
	}
	defer CloseFd(parentFd)

	fd, err := unix.Openat2(parentFd, name, &unix.OpenHow{
		Flags: uint64(
			flags |
				unix.O_CLOEXEC |
				unix.O_NOFOLLOW |
				unix.O_NONBLOCK,
		),

		Mode:    mode,
		Resolve: unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		return -1, nil, ErrFileOpen
	}

	defer func() {
		if !keep {
			CloseFd(fd)
		}
	}()

	if err := unix.Fstat(fd, stat); err != nil {
		return -1, nil, ErrStat
	}

	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return -1, nil, ErrIsNotReg
	}

	keep = true

	return fd, stat, nil
}

// Unlink unlinks given path.
func Unlink(path string, dir bool) error {
	if err := HasDotDots(path); err != nil {
		return err
	}

	var (
		clean  = filepath.Clean(path)
		parent = filepath.Dir(clean)
		name   = filepath.Base(clean)
		flags  int

		how = &unix.OpenHow{
			Flags: unix.O_PATH |
				unix.O_CLOEXEC |
				unix.O_NOFOLLOW,
			Resolve: unix.RESOLVE_NO_SYMLINKS,
		}

		stat = &unix.Stat_t{}
		cmp  = &unix.Stat_t{}
	)

	parentFd, err := unix.Openat2(
		unix.AT_FDCWD,
		parent,
		how,
	)
	if err != nil {
		return ErrFileOpen
	}
	defer CloseFd(parentFd)

	fd, err := unix.Openat2(
		parentFd,
		name,
		how,
	)
	if err != nil {
		return ErrFileOpen
	}
	defer CloseFd(fd)

	if err := unix.Fstat(fd, stat); err != nil {
		return ErrStat
	}

	mode := stat.Mode & unix.S_IFMT

	switch {
	case dir && mode != unix.S_IFDIR:
		return ErrIsNotDir

	case dir:
		flags = unix.AT_REMOVEDIR

	case mode != unix.S_IFREG:
		return ErrIsNotReg
	}

	if err := unix.Fstatat(parentFd, name, cmp, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return ErrStat
	}

	if cmp.Dev != stat.Dev || cmp.Ino != stat.Ino {
		if dir {
			return ErrIsNotDir
		}

		return ErrIsNotReg
	}

	if err := unix.Unlinkat(parentFd, name, flags); err != nil {
		return ErrFileDel
	}

	return nil
}

// Close closes given file includes logging.
func Close(closer io.Closer) {
	if err := closer.Close(); err != nil {
		name := ""

		if fn, ok := closer.(interface{ Name() string }); ok {
			name = fn.Name()
		}

		slog.Error(ErrFileClose.Error(), "msg", err, "type", fmt.Sprintf("%T", closer), "name", name)
	}
}

// CloseFd closes given fd includes logging. Refuses 0, 1 and 2 stdin / stdout / stderr.
func CloseFd(fd int) {
	if fd < 3 {
		return
	}

	if err := unix.Close(fd); err != nil {
		slog.Error(ErrFileClose.Error(), "msg", err, "fd", fd)
	}
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
	curr := &unix.Stat_t{}

	if err := unix.Lstat(path, curr); err != nil {
		return "", fmt.Errorf("%w, %v, %v", ErrStat, err, path)
	}

	for path != "/" {
		var (
			parent = filepath.Dir(path)
			cmp    = &unix.Stat_t{}
		)

		if err := unix.Lstat(parent, cmp); err != nil {
			return "", fmt.Errorf("%w, %v, %v", ErrStat, err, parent)
		}

		if cmp.Dev != curr.Dev {
			return path, nil
		}

		path = parent
		curr = cmp
	}

	return "/", nil
}

// MvAt moves given src to given dst confined by given root.
func MvAt(root *os.Root, src string, dst string, attr *Attr) error {
	if root == nil {
		return ErrPathRoot
	}

	src, err := RootName(root, src)
	if err != nil {
		return err
	}

	dst, err = RootName(root, dst)
	if err != nil {
		return err
	}

	if src == dst {
		return nil
	}

	info, err := root.Lstat(src)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrStat, err, src)
	}

	if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%w, %v", ErrIsNotReg, src)
	}

	if err := copyAt(root, src, dst, attr); err != nil {
		return err
	}

	if err := root.Remove(src); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileDel, err, src)
	}

	return nil
}

// copyAt copies src to dst within root.
func copyAt(root *os.Root, src string, dst string, attr *Attr) error {
	srcFile, err := root.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, src)
	}
	defer Close(srcFile)

	srcStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrStat, err, src)
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%w, %v", ErrIsNotReg, src)
	}

	mode := srcStat.Mode().Perm()
	if attr != nil {
		mode = attr.Mode.Perm()
	}

	dstFile, err := root.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, dst)
	}

	var keep bool

	defer func() {
		Close(dstFile)

		if !keep {
			if err := root.Remove(dst); err != nil {
				slog.Error(ErrFileDel.Error(), "msg", err)
			}
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileCopy, err, dst)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileSync, err, dst)
	}

	if err := applyAttrAt(root, dst, attr); err != nil {
		return err
	}

	keep = true

	return nil
}

// applyAttrAt applies attr to name within root. No-op when attr is nil.
func applyAttrAt(root *os.Root, name string, attr *Attr) error {
	if attr == nil {
		return nil
	}

	if err := root.Chmod(name, attr.Mode.Perm()); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrChmod, err, name)
	}

	if err := root.Chown(name, attr.UID, attr.GID); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrChown, err, name)
	}

	return nil
}

// MvToRoot moves given src to given dstRoot. src may be outside dstRoot.
// Opened via Openat2 as RESOLVE_NO_SYMLINKS / O_NOFOLLOW.
func MvToRoot(dstRoot *os.Root, src string, dst string, attr *Attr) error {
	if dstRoot == nil {
		return ErrPathRoot
	}

	if err := HasDotDots(src); err != nil {
		return err
	}

	dstName, err := RootName(dstRoot, dst)
	if err != nil {
		return err
	}

	src = filepath.Clean(src)

	var (
		srcParent = filepath.Dir(src)
		srcName   = filepath.Base(src)
	)

	srcParentFd, err := openDir(srcParent)
	if err != nil {
		return err
	}
	defer CloseFd(srcParentFd)

	srcFd, err := unix.Openat2(srcParentFd, srcName, &unix.OpenHow{
		Flags: uint64(
			unix.O_RDONLY |
				unix.O_NONBLOCK |
				unix.O_CLOEXEC |
				unix.O_NOFOLLOW),

		Resolve: uint64(unix.RESOLVE_NO_SYMLINKS),
	})
	if err != nil {
		return ErrFileOpen
	}
	defer CloseFd(srcFd)

	stat := &unix.Stat_t{}

	if err := unix.Fstat(srcFd, stat); err != nil {
		return ErrStat
	}

	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return ErrIsNotReg
	}

	mode := fs.FileMode(stat.Mode & 0o777)

	if attr != nil {
		mode = attr.Mode.Perm()
	}

	if err := dstRoot.MkdirAll(filepath.Dir(dstName), 0o700); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrDirCreate, err, dst)
	}

	dstFile, err := dstRoot.OpenFile(dstName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, dst)
	}

	keep := false
	defer func() {
		Close(dstFile)

		if !keep {
			if err := dstRoot.Remove(dstName); err != nil {
				slog.Error(ErrFileDel.Error(), "msg", err)
			}
		}
	}()

	dstFd := int(dstFile.Fd())

	if err := copyFd(srcFd, dstFd); err != nil {
		return ErrFileCopy
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileSync, err, dst)
	}

	if err := applyAttrAt(dstRoot, dstName, attr); err != nil {
		return err
	}

	keep = true

	if err := unix.Unlinkat(srcParentFd, srcName, 0); err != nil {
		return ErrFileDel
	}

	return nil
}

// Mv moves given src to dst. For this proj it copies to support xmnts.
func Mv(src string, dst string, attr *Attr) error {
	if err := HasDotDots(src, dst); err != nil {
		return err
	}

	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	var (
		srcParent = filepath.Dir(src)
		dstParent = filepath.Dir(dst)
		srcName   = filepath.Base(src)
		dstName   = filepath.Base(dst)
	)

	srcStat := &unix.Stat_t{}

	if err := unix.Lstat(src, srcStat); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrStat, err, src)
	}

	if srcStat.Mode&unix.S_IFMT == unix.S_IFDIR {
		return fmt.Errorf("%w, %v", ErrIsDir, src)
	}

	if src == dst {
		return nil
	}

	if err := os.MkdirAll(dstParent, 0o700); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrDirCreate, err, dstParent)
	}

	srcParentFd, err := openDir(srcParent)
	if err != nil {
		return err
	}
	defer CloseFd(srcParentFd)

	dstParentFd, err := openDir(dstParent)
	if err != nil {
		return err
	}
	defer CloseFd(dstParentFd)

	return copyRemoveFile(srcParentFd, srcName, dstParentFd, dstName, attr)
}

func copyRemoveFile(srcParentFd int, srcPath string, dstParentFd int, dstPath string, attr *Attr) error {
	var (
		closed bool
		keep   bool

		how = &unix.OpenHow{
			Flags: uint64(
				unix.O_RDONLY |
					unix.O_NONBLOCK |
					unix.O_CLOEXEC |
					unix.O_NOFOLLOW),

			Resolve: uint64(unix.RESOLVE_NO_SYMLINKS),
		}

		stat = &unix.Stat_t{}
	)

	srcFd, err := unix.Openat2(srcParentFd, srcPath, how)
	if err != nil {
		return ErrFileOpen
	}
	defer CloseFd(srcFd)

	switch {
	case unix.Fstat(srcFd, stat) != nil:
		return ErrStat

	case stat.Mode&unix.S_IFMT != unix.S_IFREG:
		return ErrIsNotReg
	}

	mode := stat.Mode & 0o777

	if attr != nil {
		mode = uint32(attr.Mode.Perm())
	}

	how = &unix.OpenHow{
		Flags: uint64(
			unix.O_WRONLY |
				unix.O_CREAT |
				unix.O_TRUNC |
				unix.O_CLOEXEC |
				unix.O_NOFOLLOW,
		),

		Mode:    uint64(mode),
		Resolve: uint64(unix.RESOLVE_NO_SYMLINKS),
	}

	dstFd, err := unix.Openat2(dstParentFd, dstPath, how)
	if err != nil {
		return ErrFileOpen
	}

	defer func() {
		if !closed {
			CloseFd(dstFd)
		}

		if !keep {
			if err := unix.Unlinkat(dstParentFd, dstPath, 0); err != nil {
				slog.Error(ErrFileDel.Error(), "msg", err)
			}
		}
	}()

	if err := copyFd(srcFd, dstFd); err != nil {
		return ErrFileCopy
	}

	if attr != nil {
		if err := unix.Fchown(dstFd, attr.UID, attr.GID); err != nil {
			return ErrFileCopy
		}

		if err := unix.Fchmod(dstFd, uint32(attr.Mode.Perm())); err != nil {
			return ErrFileCopy
		}
	} else {
		if err := unix.Fchmod(dstFd, stat.Mode&0o777); err != nil {
			return ErrFileCopy
		}
	}

	switch {
	case unix.Fsync(dstFd) != nil:
		return ErrFileCopy

	case unix.Close(dstFd) != nil:
		closed = true

		return ErrFileCopy
	}

	closed = true
	keep = true

	if err := unix.Unlinkat(srcParentFd, srcPath, 0); err != nil {
		return ErrFileDel
	}

	return nil
}

func openDir(path string) (int, error) {
	how := &unix.OpenHow{
		Flags: uint64(
			unix.O_RDONLY |
				unix.O_DIRECTORY |
				unix.O_CLOEXEC,
		),

		Resolve: uint64(unix.RESOLVE_NO_SYMLINKS),
	}

	fd, err := unix.Openat2(unix.AT_FDCWD, path, how)
	if err != nil {
		return -1, ErrFileOpen
	}

	return fd, nil
}

func copyFd(srcFd int, dstFd int) error {
	buff := make([]byte, 8<<20)

	for {
		pos, err := unix.Read(srcFd, buff)
		if pos > 0 {
			total := 0

			for total < pos {
				wrote, werr := unix.Write(dstFd, buff[total:pos])

				switch {
				case wrote > 0:
					total += wrote

				case werr == unix.EINTR:
					continue

				case werr != nil:
					return werr

				// no progress but no err, issue io err rather
				// than loop forever.
				case wrote == 0:
					return unix.EIO
				}
			}
		}

		switch {
		case err == unix.EINTR:
			continue

		case err != nil:
			return err

		case pos == 0:
			return nil
		}
	}
}
