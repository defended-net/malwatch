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

func TestReadTOML(t *testing.T) {
	var (
		path = filepath.Join(t.TempDir(), t.Name())

		toml = `[example]
  key = "value"`
	)

	if err := os.WriteFile(path, []byte(toml), 0600); err != nil {
		t.Fatalf("unable to write to file: %v", err)
	}

	var cfg struct {
		Example struct {
			Key string `toml:"key"`
		} `toml:"example"`
	}

	if err := ReadTOML(path, &cfg); err != nil {
		t.Fatalf("read toml error: %v", err)
	}

	if cfg.Example.Key != "value" {
		t.Errorf("unexpected read toml result: got %v, want %v", cfg.Example.Key, "value")
	}
}

func TestInstallTOML(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if err := InstallTOML(input, struct{}{}); err != nil {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLExists(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	file, err := os.Create(input)
	if err != nil {
		t.Fatalf("toml write error: %v", err)
	}
	defer file.Close()

	if err := InstallTOML(input, &struct{}{}); !errors.Is(err, fs.ErrExist) {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLDisabled(t *testing.T) {
	var (
		path     = filepath.Join(t.TempDir(), t.Name()+".toml")
		disabled = strings.Replace(path, ".toml", ".disabled", 1)
	)

	file, err := os.Create(disabled)
	if err != nil {
		t.Fatalf("disabled toml write error: %v", err)
	}
	defer file.Close()

	if err := InstallTOML(path, struct{}{}); err != nil {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLErrs(t *testing.T) {
	input := "/dev/null/" + t.Name()

	if err := InstallTOML(input, struct{}{}); !errors.Is(err, ErrTOMLRead) {
		t.Errorf("unexpected install toml error: %v", err)
	}
}

func TestWriteTOML(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if err := WriteTOML(input, struct{}{}); err != nil {
		t.Errorf("write toml error: %v", err)
	}
}

func TestWriteTOMLErrs(t *testing.T) {
	input := "//"

	if err := WriteTOML(input, struct{}{}); !errors.Is(err, ErrFileOpen) {
		t.Errorf("unexpected write toml error: %v", err)
	}
}

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
		CTime: time.Unix(stat.Ctim.Sec, 0).UTC(),
		MTime: time.Unix(stat.Mtim.Sec, 0).UTC(),
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
			want:  ErrDirMv,
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
	path := "/dev/null/" + t.Name()

	attr := &Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	if err := Mv(path, path, attr); !errors.Is(err, ErrFileOpen) {
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
	result, err := WalkByExt(t.Name(), ".ext")
	if len(result) != 0 {
		t.Fatalf("unexpected walk by ext result %v", result)
	}

	if !errors.Is(err, ErrWalk) {
		t.Errorf("unexpected walk by ext error: %v, want %v", err, ErrWalk)
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
			want:  ErrPathNotAbs,
		},

		"rel-prefix-abs": {
			input: "/../",
			want:  ErrPathRoot,
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
	input := t.Name()

	if _, err := MntPoint(input); !errors.Is(err, ErrStat) {
		t.Errorf("unexpected get mnt point error: %v", err)
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

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create error: %s", err)
	}

	if _, result := IsExpired(time.Now().Add(time.Hour*24), input); result != true {
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

			if _, result := IsExpired(time.Now().Add(-(time.Hour * 24)), path); result != test.want {
				t.Errorf("unexpected is expired timestomp result %v, want %v", result, test.want)
			}
		})
	}
}
