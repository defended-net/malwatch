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
		dir    = t.TempDir()
		sample = filepath.Join(dir, "file.php")
	)

	if err := os.WriteFile(sample, []byte(`hello-world`), 0600); err != nil {
		t.Errorf("file write error: %v", err)
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
				dir,
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
		t.Errorf("env mock error: %v", err)
	}

	env.Cfg.Scans.Paths = []string{dir}

	if err := sig.Refresh(env); err != nil {
		t.Fatalf("sig refresh error: %s", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if result := Do(env, test.input); result != test.want {
				t.Errorf("scan start error: %v", result)
			}
		})
	}
}

func TestDoErrs(t *testing.T) {
	input := []string{
		"target/file.php",
	}

	if err := Do(nil, input); !errors.Is(err, fsys.ErrPathNotAbs) {
		t.Errorf("unexpected scan start error: %v", err)
	}
}
