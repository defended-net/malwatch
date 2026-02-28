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

func TestNewAttr(t *testing.T) {
	file, err := os.OpenFile(filepath.Join(t.TempDir(), t.Name()), os.O_CREATE, 0600)
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat error: %v", err)
	}

	result := NewAttr(stat)

	want := &Attr{
		UID:   int(stat.Uid),
		GID:   int(stat.Gid),
		Mode:  fs.FileMode(stat.Mode).Perm(),
		CTime: time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC(),
		MTime: time.Unix(stat.Mtim.Sec, stat.Ctim.Nsec).UTC(),
	}

	if !reflect.DeepEqual(result, want) {
		t.Fatalf("unexpected attr: %v, want %v", result, want)
	}
}

func TestQuarantinePath(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = "/tmp/file"
		got  = QuarantinePath(dir, path)
	)

	if !strings.HasPrefix(got, dir) {
		t.Errorf("unexpected quarantine path: got %v, want prefix %v", got, dir)
	}
}

func TestMv(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}

	attr := &Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	if err := Mv(file.Name(), filepath.Join(t.TempDir(), t.Name()), attr); err != nil {
		t.Errorf("mv error: %v", err)
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

		"relative": {
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

			if err := Mv(test.input, test.input, attr); !errors.Is(err, test.want) {
				t.Errorf("unexpected mv error: %v", err)
			}
		})
	}
}

func TestMvErrs(t *testing.T) {
	var (
		input = "/dev/null/" + t.Name()
		want  = ErrStat
	)

	if err := Mv(input, input, &Attr{}); !errors.Is(err, want) {
		t.Errorf("unexpected mv error: %v, want %v", err, "not a directory")
	}
}

func TestWalk(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, t.Name())

		want = []string{
			path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file open error: %v", err)
	}

	result, err := Walk(dir)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	if !slices.Equal(result, []string{path}) {
		t.Fatalf("unexpected walk result %v, want %v", result, want)
	}
}

func TestWalkErrs(t *testing.T) {
	if _, err := Walk(t.Name()); !errors.Is(err, ErrWalk) {
		t.Errorf("unexpected walk error: %v", err)
	}
}

func TestWalkByExt(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, t.Name()+".ext")

		want = []string{
			path,
		}
	)

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file open error: %v", err)
	}

	if _, err := os.Create(path + ".txe"); err != nil {
		t.Fatalf("file open error: %v", err)
	}

	result, err := WalkByExt(dir, ".ext")
	if err != nil {
		t.Fatalf("walk by ext error: %v", err)
	}

	if !slices.Equal(result, []string{path}) {
		t.Errorf("unexpected walk by ext result %v, want %v", result, want)
	}
}

func TestWalkByExtErrs(t *testing.T) {
	want := ErrWalk

	result, err := WalkByExt(t.Name(), ".ext")
	if len(result) != 0 {
		t.Fatalf("unexpected walk by ext result %v", result)
	}

	if !errors.Is(err, want) {
		t.Errorf("unexpected walk by ext error: %v, want %v", err, want)
	}
}

func TestGetMntPoint(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"root": {
			input: "/",
			want:  "/",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := MntPoint(test.input)
			if err != nil {
				t.Fatalf("get mnt point error: %v", err)
			}

			if result != test.want {
				t.Errorf("unexpected get mnt point result %v, want %v", result, test.want)
			}
		})
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

		"rel-prefix": {
			input: "../",
			want:  ErrPathTravers,
		},

		"rel-prefix-abs": {
			input: "/../",
			want:  ErrPathTravers,
		},

		"zero-length": {
			input: "",
			want:  ErrPathNotAbs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := HasDotDots(test.input); !errors.Is(err, test.want) {
				t.Errorf("unexpected has dot dots result %v, want %v", err, test.want)
			}
		})
	}
}

func TestGetMntPointErrs(t *testing.T) {
	var (
		input = t.Name()
		want  = ErrStat
	)

	if _, got := MntPoint(input); !errors.Is(got, want) {
		t.Errorf("unexpected get mnt point error: %v", got)
	}
}

func TestGetQuarantinePath(t *testing.T) {
	var (
		dir  = t.TempDir()
		want = filepath.Join(dir, t.Name())
		got  = QuarantinePath(dir, want)
	)

	if !strings.HasPrefix(got, dir) {
		t.Errorf("unexpected quarantine path: %v, want %v", got, want)
	}
}

func TestIsRel(t *testing.T) {
	tests := map[string]struct {
		input string
		path  string
		want  bool
	}{
		"relative": {
			input: "/test/test/a.ext",
			path:  "/test/test",
			want:  true,
		},

		"not-relative": {
			input: "/test/test-b/a.ext",
			path:  "/test/test",
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if result := IsRel(test.input, test.path); result != test.want {
				t.Errorf("unexpected is relative result %v, want %v", result, test.want)
			}
		})
	}
}

func TestIsRelErrs(t *testing.T) {
	if result := IsRel("/a", "./b/c"); result != false {
		t.Errorf("unexpected is relative result %v, want %v", result, false)
	}
}

func TestIsExpired(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	file, err := os.Create(input)
	if err != nil {
		t.Fatalf("file create error: %s", err)
	}
	defer file.Close()

	if result, _ := IsExp(time.Now().Add(time.Hour*24), int(file.Fd())); result != true {
		t.Errorf("unexpected is expired result %v, want %v", result, true)
	}
}

func TestIsExpiredTimestomp(t *testing.T) {
	tests := map[string]struct {
		input time.Time
		want  bool
	}{
		"under": {
			input: time.Now(),
			want:  false,
		},

		"over": {
			input: time.Now().Add(-(time.Hour * 25)),
			want:  false,
		},
	}

	path := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create error: %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := os.Chtimes(path, time.Now(), test.input); err != nil {
				t.Errorf("mtime alert error: %s", err)
			}

			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("file open error: %s", err)
			}
			defer file.Close()

			if result, _ := IsExp(time.Now().Add(-(time.Hour * 24)), int(file.Fd())); result != test.want {
				t.Errorf("unexpected is expired timestomp result %v, want %v", result, test.want)
			}
		})
	}
}
