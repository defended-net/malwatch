// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
)

var sample = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`

func TestNew(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock error %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error %v", err)
	}

	tests := map[string]struct {
		input string
		want  *state.Job
	}{
		"detect": {
			input: sample,

			want: &state.Job{
				Hits: make(chan *state.Hit),
			},
		},

		"no-detect": {
			input: `hello-world`,

			want: &state.Job{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.input), 0600); err != nil {
				t.Fatalf("file write error %v", err)
			}

			scan, err := New(env, tmp)
			if err != nil {
				t.Fatalf("create scan error %v", err)
			}

			if err := scan.Run(); err != nil {
				t.Errorf("run scan error %v", err)
			}
		})
	}
}

func TestGetScanPaths(t *testing.T) {
	want := t.TempDir()

	input, err := Glob([]string{want})
	if err != nil {
		t.Fatalf("glob err %v", err)
	}

	if got := input[0]; got != want {
		t.Errorf("unexpected glob result %v, want %v", got, want)
	}
}

func TestGlobErrs(t *testing.T) {
	if _, got := Glob([]string{"["}); got == nil {
		t.Errorf("unexpected glob success")
	}
}

func TestGlobNoPaths(t *testing.T) {
	want := ErrNoScanPaths

	if _, got := Glob([]string{filepath.Join(t.TempDir(), "noexist*")}); !errors.Is(got, want) {
		t.Errorf("unexpected glob err %v, want %v", got, want)
	}
}

func TestGroupInvalidPath(t *testing.T) {
	got := Group([]string{filepath.Join(t.TempDir(), "not-exist")})

	if len(got) != 0 {
		t.Errorf("unexpected group result %v, want empty", got)
	}
}

func TestGroupFile(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if _, err := os.Create(input); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if got := Group([]string{input}); len(got) == 0 {
		t.Errorf("unexpected empty group result")
	}
}

func TestRefresh(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	scan, err := New(env, t.TempDir())
	if err != nil {
		t.Fatalf("scan create err %v", err)
	}

	if err := scan.refresh(123); err != nil {
		t.Errorf("refresh err %v", err)
	}

	if scan.rev != 123 {
		t.Errorf("unexpected rev %v, want %v", scan.rev, 123)
	}
}

func TestGlob(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"ok": {
			input: []string{t.TempDir()},
			want:  nil,
		},

		"compound": {
			input: []string{t.TempDir(), t.TempDir()},
			want:  nil,
		},

		"no-match": {
			input: []string{filepath.Join(t.TempDir(), "not-exist-*")},
			want:  ErrNoScanPaths,
		},

		"pattern-invalid": {
			input: []string{"["},
			want:  filepath.ErrBadPattern,
		},

		"empty": {
			input: []string{},
			want:  ErrNoScanPaths,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, got := Glob(test.input)

			if !errors.Is(got, test.want) {
				t.Errorf("unexpected glob err %v, want %v", got, test.want)
			}
		})
	}
}

func TestGroup(t *testing.T) {
	var (
		tmp   = t.TempDir()
		input = filepath.Join(t.TempDir(), t.Name())
	)

	if err := os.WriteFile(input, []byte(t.Name()), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	tests := map[string]struct {
		input     []string
		wantDirs  bool
		wantFiles bool
		wantLen   int
	}{
		"dir": {
			input:    []string{tmp},
			wantDirs: true,
			wantLen:  1,
		},

		"file": {
			input:     []string{input},
			wantFiles: true,
			wantLen:   1,
		},

		"compound": {
			input:     []string{tmp, input},
			wantDirs:  true,
			wantFiles: true,
			wantLen:   1,
		},

		"not-exist": {
			input:   []string{filepath.Join(t.TempDir(), "not-exist")},
			wantLen: 0,
		},

		"empty": {
			input:   []string{},
			wantLen: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Group(test.input)

			if len(got) != test.wantLen {
				t.Errorf("unexpected group len %v, want %v", len(got), test.wantLen)
			}

			for _, paths := range got {
				if test.wantDirs && len(paths.Dirs) == 0 {
					t.Errorf("expected dirs, got none")
				}

				if test.wantFiles && len(paths.Files) == 0 {
					t.Errorf("expected files, got none")
				}
			}
		})
	}
}

func TestGroupSym(t *testing.T) {
	var (
		tmp   = t.TempDir()
		input = filepath.Join(t.TempDir(), t.Name())
	)

	if err := os.Symlink(tmp, input); err != nil {
		t.Fatalf("symlink err %v", err)
	}

	if got := Group([]string{input}); len(got) != 0 {
		t.Errorf("expected no sym, got %v", len(got))
	}
}
