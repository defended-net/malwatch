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
		t.Fatalf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Do(env, test.input); !errors.Is(got, test.want) {
				t.Fatalf("unexpected do result %v, want %v", got, test.want)
			}
		})
	}
}

func TestDoEmptyStatus(t *testing.T) {
	var (
		input = &hit.History{
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

		want = ErrStatusUnknown
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	if _, err := os.Create(filepath.Join(t.TempDir(), t.Name())); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if err := input.Save(env.Db); err != nil {
		t.Fatalf("db save err %v", err)
	}

	if got := Do(env, []string{"/target/test-restore.php"}); !errors.Is(got, want) {
		t.Fatalf("unexpected do err %v, want %v", got, want)
	}
}

func TestDoS3(t *testing.T) {
	var (
		input = &hit.History{
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

		want = fsys.ErrFileOpen
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if env.Cfg.Secrets.S3.Endpoint == "" {
		t.Skip("s3 secret not stored")
	}

	if env.Cfg.Secrets.S3.Endpoint == "" {
		t.Skip("s3 secret not stored")
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	if err := input.Save(env.Db); err != nil {
		t.Fatalf("db save err %v", err)
	}

	if got := Do(env, []string{"/target/test-restore.php"}); !errors.Is(got, want) {
		t.Fatalf("unexpected do err %v, want %v", got, want)
	}
}

func TestDoLocal(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	var (
		quarDir = env.Cfg.Acts.Quarantine.Dir
		dst     = filepath.Join(t.TempDir(), "test-restore.php")
		quar    = filepath.Join(quarDir, "test-restore.php")

		input = &hit.History{
			Target: "fs",

			Paths: hit.Paths{
				dst: {
					{
						Rules:  []string{"eicar"},
						Status: quar,
						Attr:   &fsys.Attr{Mode: 0600},
					},
				},
			},
		}
	)

	if err := os.MkdirAll(quarDir, 0700); err != nil {
		t.Fatalf("mkdir err %v", err)
	}

	if err := os.WriteFile(quar, []byte(t.Name()), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if err := input.Save(env.Db); err != nil {
		t.Fatalf("db save err %v", err)
	}

	// Restore copies quarantine file back to dst.
	_ = Do(env, []string{dst})
}
