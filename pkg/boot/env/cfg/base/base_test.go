// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package base

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/scan"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
)

func TestNew(t *testing.T) {
	var (
		want = t.TempDir()

		input = &path.Paths{
			Install: &path.Install{},

			Cfg: &path.Cfg{
				Base: want,
			},
		}

		got = New(input)
	)

	if got.path != want {
		t.Errorf("unexpected path result %v, want %v", got.path, want)
	}
}

func TestLoad(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())

		mock = `Identifier = ""
Cores = 1
Threads = 1

[Scans]
  Targets = ["^/var/www/(?P<target>[^/]+)"]
  Paths = ["/var/www/*"]
  Timeout = 60
  MaxAge = 0
  BlkSz = 65536
  BatchSz = 500
  [Scans.Monitor]
    Timeout = 5

[Database]
  Dir = "/tmp"

[Log]
  Dir = "/tmp"
  Verbose = false
`
	)

	if err := os.WriteFile(path, []byte(mock), 0600); err != nil {
		t.Errorf("file write err %v", err)
	}

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	cfg := &Cfg{
		path: path,
	}

	if got := cfg.Load(root); got != nil {
		t.Errorf("load err %v", got)
	}
}

func TestPath(t *testing.T) {
	var (
		want = t.Name()

		input = &Cfg{
			path: want,
		}
	)

	if got := input.Path(); got != t.Name() {
		t.Errorf("unexpected path result %v, want %v", got, want)
	}
}

func TestIdentifier(t *testing.T) {
	input := &Cfg{
		Identifier: "",

		Scans: &scan.Cfg{
			Targets: []string{
				`^/(?P<target>target)/?(.*)`,
			},
		},
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Errorf("hostname err %v", err)
	}

	if err := input.Validate(); err != nil {
		t.Errorf("validate err %v", err)
	}

	if input.Identifier != hostname {
		t.Errorf("unexpected identifier result %v, want %v", input.Identifier, hostname)
	}
}

func TestValidateErr(t *testing.T) {
	var (
		input = &Cfg{
			Identifier: "test",

			Scans: &scan.Cfg{
				Targets: []string{
					`[`,
				},
			},
		}

		want = ErrReTarget
	)

	if got := input.Validate(); !errors.Is(got, want) {
		t.Errorf("unexpected validate err %v, want %v", got, want)
	}
}

func TestMock(t *testing.T) {
	var (
		tmp = t.TempDir()

		paths = &path.Paths{
			Install: &path.Install{
				Log: filepath.Join(tmp, t.Name()+".log"),
				Tmp: filepath.Join(tmp, "tmp"),
			},

			Cfg: &path.Cfg{
				Dir:     tmp,
				Base:    filepath.Join(tmp, "cfg", "cfg.toml"),
				Secrets: filepath.Join(tmp, "cfg", "secrets.toml"),
				Acts:    filepath.Join(tmp, "cfg", "actions.toml"),
			},

			Plat: &path.Plat{
				Dir: filepath.Join(tmp, "cfg", "plat"),
			},

			Alerts: &path.Alerts{
				Dir: filepath.Join(tmp, "cfg", "alerts"),
			},

			Sigs: &path.Sigs{
				Dir: filepath.Join(tmp, "sigs"),
				Src: filepath.Join(tmp, "sigs", "src"),
				Tmp: filepath.Join(tmp, "sigs", "tmp"),
			},
		}
	)

	if _, got := Mock(paths, tmp); got != nil {
		t.Errorf("cfg mock err %v", got)
	}
}
