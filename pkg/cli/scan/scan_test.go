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
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/sig"
)

func TestDo(t *testing.T) {
	var (
		tmp    = t.TempDir()
		sample = filepath.Join(tmp, "file.php")
	)

	if err := os.WriteFile(sample, []byte(t.Name()), 0600); err != nil {
		t.Errorf("file write err %v", err)
	}

	tests := map[string]struct {
		input []string
		want  error
	}{
		"all": {
			input: []string{},

			want: nil,
		},

		"dir": {
			input: []string{
				tmp,
			},

			want: nil,
		},

		"file": {
			input: []string{
				sample,
			},

			want: nil,
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	env.Cfg.Scans.Paths = []string{tmp}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Do(env, test.input); got != test.want {
				t.Errorf("do err %v", got)
			}
		})
	}
}

func TestDoErrs(t *testing.T) {
	var (
		input = []string{
			"target/file.php",
		}

		want = fsys.ErrPathNotAbs
	)

	if got := Do(nil, input); !errors.Is(got, want) {
		t.Errorf("unexpected do err %v, want %v", got, want)
	}
}
