// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package restore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/client/s3"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestDo(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"empty": {
			input: []string{
				"",
			},

			want: fsys.ErrPathNotAbs,
		},

		"not-abs": {
			input: []string{
				"dir/file",
			},

			want: fsys.ErrPathNotAbs,
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Do(env, test.input); !errors.Is(err, test.want) {
				t.Fatalf("unexpected restore result %v, want %v", err, test.want)
			}
		})
	}
}

func TestDoEmptyStatus(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %v", err)
	}

	if _, err := os.Create(filepath.Join(t.TempDir(), t.Name())); err != nil {
		t.Fatalf("file write error: %v", err)
	}

	hits := &hit.History{
		Target: "target",

		Paths: hit.Paths{
			"/target/test-restore.php": {
				{
					Rules:  []string{"eicar"},
					Status: "",
					Attr:   &fsys.Attr{},
				},
			},
		},
	}

	if err := hits.Save(env.Db); err != nil {
		t.Fatalf("db save error: %v", err)
	}

	if err := Do(env, []string{"/target/test-restore.php"}); !errors.Is(err, ErrStatusUnknown) {
		t.Fatalf("unexpected restore error %v, want %v", err, ErrStatusUnknown)
	}
}

func TestDoS3(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %v", err)
	}

	hits := &hit.History{
		Target: "fs",

		Paths: hit.Paths{
			"/target/test-restore.php": {
				{
					Rules:  []string{},
					Status: s3.Scheme + "test-restore.php",
					Attr:   &fsys.Attr{},
				},
			},
		},
	}

	if err := hits.Save(env.Db); err != nil {
		t.Fatalf("db save error: %v", err)
	}

	if err := Do(env, []string{"/target/test-restore.php"}); !errors.Is(err, fsys.ErrFileOpen) {
		t.Fatalf("unexpected restore error %v, want %v", err, fsys.ErrFileOpen)
	}
}
