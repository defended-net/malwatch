// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestOpen(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create err %v", err)
	}

	fd, stat, got := Open(input)
	if got != nil {
		t.Errorf("file open err %v", got)
	}
	defer CloseFd(fd)

	if stat == nil {
		t.Errorf("nil stat")
	}
}

func TestOpenErrs(t *testing.T) {
	want := ErrPathTraverse

	if _, _, got := Open("/../etc/file"); !errors.Is(got, want) {
		t.Errorf("unexpected file open err %v, want %v", got, want)
	}
}

func TestOpenFile(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	fd, _, got := OpenFile(input, unix.O_WRONLY|unix.O_CREAT)
	if got != nil {
		t.Errorf("open file err %v", got)
	}
	defer CloseFd(fd)
}

func TestOpenFileErrs(t *testing.T) {
	tests := map[string]struct {
		input func(t *testing.T) string
		want  error
	}{
		"not-exist": {
			input: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "not-exist", "file")
			},

			want: ErrFileOpen,
		},

		"dir": {
			input: func(t *testing.T) string {
				return t.TempDir()
			},

			want: ErrIsNotReg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, got := OpenFile(test.input(t), unix.O_RDONLY); !errors.Is(got, test.want) {
				t.Errorf("unexpected file open err %v, want %v", got, test.want)
			}
		})
	}
}

func TestUnlink(t *testing.T) {
	tests := map[string]struct {
		fn  func(t *testing.T) string
		dir bool
	}{
		"file": {
			fn: func(t *testing.T) string {
				input := filepath.Join(t.TempDir(), "file")

				if _, err := os.Create(input); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return input
			},

			dir: false,
		},

		"dir": {
			fn: func(t *testing.T) string {
				input := filepath.Join(t.TempDir(), "dir")

				if err := os.Mkdir(input, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				return input
			},

			dir: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				input = test.fn(t)
				want  = os.ErrNotExist
			)

			if err := Unlink(input, test.dir); err != nil {
				t.Errorf("unlink err %v", err)
			}

			if _, got := os.Stat(input); !errors.Is(got, want) {
				t.Errorf("unexpected stat err %v, want %v", got, want)
			}
		})
	}
}

func TestUnlinkErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) string
		dir  bool
		want error
	}{
		"dot-dots": {
			fn: func(_ *testing.T) string {
				return "/../etc/file"
			},

			dir:  false,
			want: ErrPathTraverse,
		},

		"not-exist": {
			fn: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "not-exist", "file")
			},

			dir:  false,
			want: ErrFileOpen,
		},

		"not-reg": {
			fn: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "dir")

				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				return path
			},

			dir:  false,
			want: ErrIsNotReg,
		},

		"not-dir": {
			fn: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "file")

				if _, err := os.Create(path); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return path
			},

			dir:  true,
			want: ErrIsNotDir,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Unlink(test.fn(t), test.dir); !errors.Is(got, test.want) {
				t.Errorf("unexpected unlink err %v, want %v", got, test.want)
			}
		})
	}
}

func TestOpenParent(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create err %v", err)
	}

	parent, fd, name, stat, got := OpenParent(input)
	if got != nil {
		t.Errorf("open parent err %v", got)
	}

	for _, fd := range []int{
		parent,
		fd,
	} {
		defer CloseFd(fd)
	}

	if name != filepath.Base(input) {
		t.Errorf("unexpected name %v, want %v", name, filepath.Base(input))
	}

	if stat == nil {
		t.Errorf("nil stat")
	}
}

func TestOpenParentErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) string
		want error
	}{
		"dot-dots": {
			fn: func(_ *testing.T) string {
				return "/../etc/file"
			},

			want: ErrPathTraverse,
		},

		"not-exist": {
			fn: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "not-exist", "file")
			},

			want: ErrFileOpen,
		},

		"not-reg": {
			fn: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "dir")

				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				return path
			},

			want: ErrIsNotReg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, _, _, got := OpenParent(test.fn(t)); !errors.Is(got, test.want) {
				t.Errorf("unexpected open parent err %v, want %v", got, test.want)
			}
		})
	}
}

func TestUnlinkAt(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create err %v", err)
	}

	parent, fd, name, _, err := OpenParent(input)
	if err != nil {
		t.Fatalf("open parent err %v", err)
	}

	for _, fd := range []int{
		parent,
		fd,
	} {
		defer CloseFd(fd)
	}

	if got := UnlinkAt(parent, name); got != nil {
		t.Errorf("unlink err %v", got)
	}

	if _, got := os.Stat(input); !errors.Is(got, os.ErrNotExist) {
		t.Errorf("unexpected stat err %v, want %v", got, os.ErrNotExist)
	}
}

func TestUnlinkAtErrs(t *testing.T) {
	var (
		fd, err = openDir(t.TempDir())
		want    = ErrFileDel
	)

	if err != nil {
		t.Fatalf("open dir err %v", err)
	}
	defer CloseFd(fd)

	if got := UnlinkAt(fd, "not-exist"); !errors.Is(got, want) {
		t.Errorf("unexpected unlink err %v, want %v", got, want)
	}
}

func TestClose(t *testing.T) {
	tests := map[string]struct {
		preClose bool
	}{
		"open": {
			preClose: false,
		},

		"closed": {
			// Logs error.
			preClose: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			file, err := os.Create(filepath.Join(t.TempDir(), "file"))
			if err != nil {
				t.Fatalf("file create err %v", err)
			}

			if test.preClose {
				if err := file.Close(); err != nil {
					t.Fatalf("file close err %v", err)
				}
			}

			Close(file)
		})
	}
}

func TestCloseFd(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	tests := map[string]struct {
		fd int
	}{
		"ok": {
			fd: int(file.Fd()),
		},

		"stdin": {
			fd: 0,
		},

		"stdout": {
			fd: 1,
		},

		"stderr": {
			fd: 2,
		},

		"invalid": {
			fd: 99999,
		},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			CloseFd(test.fd)
		})
	}
}

func TestMv(t *testing.T) {
	attr := &Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	tests := map[string]struct {
		fn   func(t *testing.T) (src, dst string)
		attr *Attr
	}{
		"reg": {
			fn: func(t *testing.T) (string, string) {
				input := filepath.Join(t.TempDir(), "src")

				if _, err := os.Create(input); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return input, filepath.Join(t.TempDir(), "dst")
			},

			attr: attr,
		},

		"match": {
			fn: func(t *testing.T) (string, string) {
				input := filepath.Join(t.TempDir(), "file")

				if _, err := os.Create(input); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return input, input
			},

			attr: attr,
		},

		"dst-par-missing": {
			fn: func(t *testing.T) (string, string) {
				input := filepath.Join(t.TempDir(), "src")

				if _, err := os.Create(input); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return input, filepath.Join(t.TempDir(), "dst", "file")
			},

			attr: attr,
		},

		"attr-missing": {
			fn: func(t *testing.T) (string, string) {
				input := filepath.Join(t.TempDir(), "src")

				if _, err := os.Create(input); err != nil {
					t.Fatalf("file create err %v", err)
				}

				return input, filepath.Join(t.TempDir(), "dst")
			},

			attr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			src, dst := test.fn(t)

			if got := Mv(src, dst, test.attr); got != nil {
				t.Errorf("mv err %v", got)
			}
		})
	}
}

func TestMvErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (src, dst string)
		attr *Attr
		want error
	}{
		"dir": {
			fn: func(t *testing.T) (string, string) {
				input := filepath.Join(t.TempDir(), "src")

				if err := os.Mkdir(input, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				return input, filepath.Join(t.TempDir(), "dst")
			},

			want: ErrIsDir,
		},

		"src-missing": {
			fn: func(t *testing.T) (string, string) {
				return filepath.Join(t.TempDir(), "src"), filepath.Join(t.TempDir(), "dst")
			},

			want: ErrStat,
		},

		"dot-dots": {
			fn: func(_ *testing.T) (string, string) {
				return "/../etc/file", "/tmp/x"
			},

			want: ErrPathTraverse,
		},

		"stat": {
			fn: func(_ *testing.T) (string, string) {
				input := filepath.Join("/dev/null", "file")

				return input, input
			},

			attr: &Attr{},
			want: ErrStat,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			src, dst := test.fn(t)

			if got := Mv(src, dst, test.attr); !errors.Is(got, test.want) {
				t.Errorf("unexpected mv err %v, want %v", got, test.want)
			}
		})
	}
}

func TestMvPathErrs(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"root": {
			input: "/",
			want:  ErrPathRoot,
		},

		"rel": {
			input: "dev/null/file",
			want:  ErrPathNotAbs,
		},

		"dir": {
			input: t.TempDir(),
			want:  ErrIsDir,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			attr := &Attr{
				UID:  os.Getuid(),
				GID:  os.Getgid(),
				Mode: 0600,
			}

			if got := Mv(test.input, test.input, attr); !errors.Is(got, test.want) {
				t.Errorf("unexpected mv err %v", got)
			}
		})
	}
}

func TestMvAt(t *testing.T) {
	var (
		dir = t.TempDir()
		src = filepath.Join(dir, "src")
		dst = filepath.Join(dir, "dst", "dst")

		attr = &Attr{
			UID:  os.Getuid(),
			GID:  os.Getgid(),
			Mode: 0600,
		}
	)

	if err := os.WriteFile(src, []byte(t.Name()), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		t.Fatalf("mkdir err %v", err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := MvToRoot(root, src, dst, attr); got != nil {
		t.Errorf("mv err %v", got)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Errorf("stat err %v", err)
	}

	if _, err := os.Stat(src); err == nil {
		t.Errorf("src exist")
	}
}

func TestMvToRootSameTarget(t *testing.T) {
	var (
		tmp   = t.TempDir()
		input = filepath.Join(tmp, "file")
	)

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create err %v", err)
	}

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := MvToRoot(root, input, input, nil); got != nil {
		t.Errorf("mv err %v", got)
	}
}

func TestMvRootErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, src, dst string)
		want error
	}{
		"nil": {
			fn: func(_ *testing.T) (*os.Root, string, string) {
				return nil, "src", "dst"
			},

			want: ErrPathRoot,
		},

		"escape": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "../escape", "dst"
			},

			want: ErrPathLocal,
		},

		"sym": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				var (
					tmp    = t.TempDir()
					target = filepath.Join(tmp, "target")
					src    = filepath.Join(tmp, "src.yar")
					dst    = filepath.Join(tmp, "dst.yar")
				)

				if err := os.WriteFile(target, []byte(t.Name()), 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				if err := os.Symlink(target, src); err != nil {
					t.Fatalf("symlink err %v", err)
				}

				root, err := os.OpenRoot(tmp)
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				t.Cleanup(func() {
					if _, err := os.Lstat(src); err != nil {
						t.Errorf("lstat err %v", err)
					}

					if _, err := os.Lstat(dst); err == nil {
						t.Errorf("dst created from sym")
					}
				})

				return root, src, dst
			},

			want: ErrIsNotReg,
		},

		"dir": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				var (
					tmp = t.TempDir()
					src = filepath.Join(tmp, "src")
					dst = filepath.Join(tmp, "dst")
				)

				if err := os.Mkdir(src, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				root, err := os.OpenRoot(tmp)
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, src, dst
			},

			want: ErrIsNotReg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, src, dst := test.fn(t)

			if got := MvToRoot(root, src, dst, nil); !errors.Is(got, test.want) {
				t.Errorf("unexpected mv err %v, want %v", got, test.want)
			}
		})
	}
}

func TestRootName(t *testing.T) {
	tests := map[string]struct {
		path func(tmp string) string
		want string
	}{
		"file": {
			path: func(tmp string) string {
				return filepath.Join(tmp, "file")
			},

			want: "file",
		},

		"tmp": {
			path: func(tmp string) string {
				return tmp
			},
		},

		"dot": {
			path: func(_ string) string {
				return "."
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()

			root, err := os.OpenRoot(tmp)
			if err != nil {
				t.Fatalf("open root err %v", err)
			}

			defer func() {
				// lint
				_ = root.Close()
			}()

			got, err := RootName(root, test.path(tmp))
			if err != nil {
				t.Fatalf("root filename err %v", err)
			}

			if test.want != "" && got != test.want {
				t.Errorf("unexpected root name %v, want %v", got, test.want)
			}
		})
	}
}

func TestRootNameErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, path string)
		want error
	}{
		"nil": {
			fn: func(t *testing.T) (*os.Root, string) {
				return nil, t.Name()
			},

			want: ErrPathRoot,
		},

		"escape": {
			fn: func(t *testing.T) (*os.Root, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "../escape"
			},

			want: ErrPathLocal,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, path := test.fn(t)

			if _, got := RootName(root, path); !errors.Is(got, test.want) {
				t.Errorf("unexpected root name err %v, want %v", got, test.want)
			}
		})
	}
}

func TestMntPoint(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"root": {
			input: "/",
			want:  "/",
		},

		// /proc normally diff fs from /
		"diff-dev": {
			input: "/proc/self",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := MntPoint(test.input)
			if err != nil {
				t.Fatalf("mnt point err %v", err)
			}

			if test.want != "" && got != test.want {
				t.Errorf("unexpected mnt point result %v, want %v", got, test.want)
			}

			if got == "" {
				t.Errorf("empty mnt point")
			}
		})
	}
}

func TestMntPointErrs(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"not-exist": {
			input: "/dev/null/not-exist",
			want:  ErrStat,
		},

		"name": {
			input: t.Name(),
			want:  ErrStat,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, got := MntPoint(test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected mnt point err %v, want %v", got, test.want)
			}
		})
	}
}

func TestNewAttr(t *testing.T) {
	file, err := os.OpenFile(filepath.Join(t.TempDir(), t.Name()), os.O_CREATE, 0600)
	if err != nil {
		t.Fatalf("file open err %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat err %v", err)
	}

	var (
		got = NewAttr(stat)

		want = &Attr{
			UID:   int(stat.Uid),
			GID:   int(stat.Gid),
			Mode:  fs.FileMode(stat.Mode).Perm(),
			CTime: time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC(),
			MTime: time.Unix(stat.Mtim.Sec, stat.Ctim.Nsec).UTC(),
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected create attr result %v, want %v", got, want)
	}
}

func TestQuarantinePath(t *testing.T) {
	tests := map[string]struct {
		dir  func(t *testing.T) string
		path func(dir string) string
	}{
		"missing-dir": {
			dir: func(_ *testing.T) string {
				return "/quarantine"
			},

			path: func(_ string) string {
				return "/file"
			},
		},

		"tmp": {
			dir: func(t *testing.T) string {
				return t.TempDir()
			},

			path: func(_ string) string {
				return "/tmp/file"
			},
		},

		"sub": {
			dir: func(t *testing.T) string {
				return t.TempDir()
			},

			path: func(dir string) string {
				return filepath.Join(dir, "file")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				tmp = test.dir(t)
				got = QuarantinePath(tmp, test.path(tmp))
			)

			switch {
			case got == "":
				t.Errorf("empty quarantine path")

			case !strings.HasPrefix(got, tmp):
				t.Errorf("unexpected quarantine path %v, want %v", got, tmp)
			}
		})
	}
}

func TestWalk(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())

		want = []string{
			path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file open err %v", err)
	}

	got, err := Walk(tmp)
	if err != nil {
		t.Fatalf("walk err %v", err)
	}

	if !slices.Equal(got, []string{path}) {
		t.Fatalf("unexpected walk result %v, want %v", got, want)
	}
}

func TestWalkErrs(t *testing.T) {
	want := ErrWalk

	if _, got := Walk(t.Name()); !errors.Is(got, want) {
		t.Errorf("unexpected walk err %v, want %v", got, want)
	}
}

func TestWalkByExt(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name()+".ext")

		want = []string{
			path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if _, err := os.Create(path + ".txe"); err != nil {
		t.Fatalf("file create err %v", err)
	}

	got, err := WalkByExt(tmp, ".ext")
	if err != nil {
		t.Fatalf("walk by ext err %v", err)
	}

	if !slices.Equal(got, []string{path}) {
		t.Errorf("unexpected walk by ext result %v, want %v", got, want)
	}
}

func TestWalkByExtErrs(t *testing.T) {
	var (
		got, err = WalkByExt(t.Name(), ".ext")
		want     = ErrWalk
	)

	if len(got) != 0 {
		t.Fatalf("unexpected walk by ext result %v", got)
	}

	if !errors.Is(err, want) {
		t.Errorf("unexpected walk by ext err %v, want %v", err, want)
	}
}

func TestHasDotDots(t *testing.T) {
	tests := map[string]struct {
		input string
		want  error
	}{
		"root": {
			input: "/",
			want:  ErrPathRoot,
		},

		"rel": {
			input: ".",
			want:  ErrPathNotAbs,
		},

		"rel-pfx": {
			input: "../",
			want:  ErrPathTraverse,
		},

		"rel-pfx-abs": {
			input: "/../",
			want:  ErrPathTraverse,
		},

		"zero-len": {
			input: "",
			want:  ErrPathNotAbs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := HasDotDots(test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected has dot dots result %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsRel(t *testing.T) {
	tests := map[string]struct {
		input string
		path  string
		want  bool
	}{
		"rel": {
			input: "/test/test/a.ext",
			path:  "/test/test",
			want:  true,
		},

		"not-relative": {
			input: "/test/test-b/a.ext",
			path:  "/test/test",
			want:  false,
		},

		"non-abs": {
			input: "/a",
			path:  "./b/c",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsRel(test.input, test.path); got != test.want {
				t.Errorf("unexpected is rel result %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsExp(t *testing.T) {
	tests := map[string]struct {
		mtime func() time.Time
		ttl   func() time.Time
		want  bool
	}{
		"exp": {
			mtime: nil,

			ttl: func() time.Time {
				return time.Now().Add(time.Hour * 24)
			},

			want: true,
		},

		"stomp-under": {
			mtime: func() time.Time {
				return time.Now()
			},

			ttl: func() time.Time {
				return time.Now().Add(-(time.Hour * 24))
			},

			want: false,
		},

		"stomp-over": {
			mtime: func() time.Time {
				return time.Now().Add(-(time.Hour * 25))
			},

			ttl: func() time.Time {
				return time.Now().Add(-(time.Hour * 24))
			},

			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			input := filepath.Join(t.TempDir(), "file")

			if _, err := os.Create(input); err != nil {
				t.Fatalf("file create err %s", err)
			}

			if test.mtime != nil {
				if err := os.Chtimes(input, time.Now(), test.mtime()); err != nil {
					t.Errorf("chtimes err %s", err)
				}
			}

			file, err := os.Open(input)
			if err != nil {
				t.Fatalf("file open err %s", err)
			}
			defer Close(file)

			if got, _ := IsExp(test.ttl(), int(file.Fd())); got != test.want {
				t.Errorf("unexpected is exp result %v, want %v", got, test.want)
			}
		})
	}
}

func TestIsExpErrs(t *testing.T) {
	got, stat := IsExp(time.Now(), -1)
	if got {
		t.Errorf("unexpected is exp result %v, want false", got)
	}

	if stat != nil {
		t.Errorf("unexpected stat %v, want nil", stat)
	}
}

func TestMvToRoot(t *testing.T) {
	attr := &Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	tests := map[string]struct {
		attr *Attr
	}{
		"basic": {
			attr: attr,
		},

		"no-attr": {
			attr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				srcDir = t.TempDir()
				dstDir = t.TempDir()
				src    = filepath.Join(srcDir, "src")
				dst    = filepath.Join(dstDir, "sub", "dst")
			)

			if err := os.WriteFile(src, []byte(t.Name()), 0600); err != nil {
				t.Fatalf("file write err %v", err)
			}

			root, err := os.OpenRoot(dstDir)
			if err != nil {
				t.Fatalf("open root err %v", err)
			}

			defer func() {
				// lint
				_ = root.Close()
			}()

			if got := MvToRoot(root, src, dst, test.attr); got != nil {
				t.Errorf("mv to root err %v", got)
			}

			if _, err := os.Stat(dst); err != nil {
				t.Errorf("stat err %v", err)
			}

			if _, err := os.Stat(src); err == nil {
				t.Errorf("src exist")
			}
		})
	}
}

func TestMvToRootErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, src, dst string)
		want error
	}{
		"nil": {
			fn: func(_ *testing.T) (*os.Root, string, string) {
				return nil, "/src", "/dst"
			},

			want: ErrPathRoot,
		},

		"dot-dots": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "/../etc/file", "dst"
			},

			want: ErrPathTraverse,
		},

		"escape": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "/tmp/src", "../escape"
			},

			want: ErrPathLocal,
		},

		"src-missing": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				input := filepath.Join(t.TempDir(), "missing")

				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, input, "dst"
			},

			want: ErrFileOpen,
		},

		"src-parent-missing": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				input := filepath.Join(t.TempDir(), "missing", "src")

				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, input, "dst"
			},

			want: ErrFileOpen,
		},

		"src-dir": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				var (
					tmp = t.TempDir()
					src = filepath.Join(tmp, "src")
				)

				if err := os.Mkdir(src, 0700); err != nil {
					t.Fatalf("mkdir err %v", err)
				}

				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, src, "dst"
			},

			want: ErrIsNotReg,
		},

		"src-sym": {
			fn: func(t *testing.T) (*os.Root, string, string) {
				var (
					tmp    = t.TempDir()
					target = filepath.Join(tmp, "target")
					src    = filepath.Join(tmp, "src")
				)

				if err := os.WriteFile(target, []byte(t.Name()), 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				if err := os.Symlink(target, src); err != nil {
					t.Fatalf("symlink err %v", err)
				}

				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, src, "dst"
			},

			want: ErrFileOpen,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, src, dst := test.fn(t)

			if got := MvToRoot(root, src, dst, nil); !errors.Is(got, test.want) {
				t.Errorf("unexpected mv to root err %v, want %v", got, test.want)
			}
		})
	}
}

func TestMvOpenDirErrs(t *testing.T) {
	var (
		tmp    = t.TempDir()
		notDir = filepath.Join(tmp, "file")
		src    = filepath.Join(notDir, "src")
		dst    = filepath.Join(t.TempDir(), "dst")
	)

	if err := os.WriteFile(notDir, []byte(t.Name()), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if got := Mv(src, dst, nil); !errors.Is(got, ErrStat) {
		t.Errorf("unexpected mv err %v, want %v", got, ErrStat)
	}
}
