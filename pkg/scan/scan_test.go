// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
)

var (
	sample = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`
)

func TestNew(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, t.Name())
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %v", err)
	}

	tests := map[string]struct {
		input string
		want  *state.Job
	}{
		"detection": {
			input: sample,

			want: &state.Job{
				Hits: make(chan *state.Hit),
			},
		},

		"no-detection": {
			input: `hello-world`,

			want: &state.Job{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.input), 0660); err != nil {
				t.Fatalf("file write error: %v", err)
			}

			scan, err := New(env, dir)
			if err != nil {
				t.Fatalf("create error: %v", err)
			}

			if err := scan.Run(); err != nil {
				t.Errorf("run error: %v", err)
			}
		})
	}
}

func TestGetScanPaths(t *testing.T) {
	want := t.TempDir()

	paths, err := Glob([]string{want})
	if err != nil {
		t.Fatalf("get scan paths error: %v", err)
	}

	got := paths[0]

	if got != want {
		t.Errorf("unexpected get scan paths result %v, want %v", got, want)
	}
}
