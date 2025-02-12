// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package base

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/scan"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
)

func TestNew(t *testing.T) {
	want := t.TempDir()

	input := &path.Paths{
		Install: &path.Install{},

		Cfg: &path.Cfg{
			Base: want,
		},
	}

	cfg := New(input)

	if cfg.path != want {
		t.Errorf("unexpected cfg path result %v, want %v", cfg.path, want)
	}
}

func TestLoad(t *testing.T) {
	mock := `Identifier = ""
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

	path := filepath.Join(t.TempDir(), t.Name())

	if err := os.WriteFile(path, []byte(mock), 0600); err != nil {
		t.Errorf("cfg write err: %v", err)
	}

	cfg := &Cfg{
		path: path,
	}

	if err := cfg.Load(); err != nil {
		t.Errorf("cfg load err: %v", err)
	}
}

func TestPath(t *testing.T) {
	want := t.Name()

	cfg := &Cfg{
		path: want,
	}

	got := cfg.Path()

	if got != t.Name() {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestIdentifier(t *testing.T) {
	cfg := &Cfg{
		Identifier: "",

		Scans: &scan.Cfg{
			Targets: []string{
				`^/(?P<target>target)/?(.*)`,
			},
		},
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Errorf("hostname lookup error %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("cfg validate error %v", err)
	}

	if cfg.Identifier != hostname {
		t.Errorf("unexpected identifier result %v, want %v", cfg.Identifier, hostname)
	}
}

func TestMock(t *testing.T) {
	dir := t.TempDir()

	paths := &path.Paths{
		Install: &path.Install{
			Log: filepath.Join(dir, t.Name()+".log"),
			Tmp: filepath.Join(dir, "tmp"),
		},

		Cfg: &path.Cfg{
			Dir:     dir,
			Base:    filepath.Join(dir, "cfg", "cfg.toml"),
			Secrets: filepath.Join(dir, "cfg", "secrets.toml"),
			Acts:    filepath.Join(dir, "cfg", "actions.toml"),
		},

		Plat: &path.Plat{
			Dir: filepath.Join(dir, "cfg", "plat"),
		},

		Alerts: &path.Alerts{
			Dir: filepath.Join(dir, "cfg", "alerts"),
		},

		Sigs: &path.Sigs{
			Dir: filepath.Join(dir, "sigs"),
			Src: filepath.Join(dir, "sigs", "src"),
			Tmp: filepath.Join(dir, "sigs", "tmp"),
		},
	}

	if _, err := Mock(paths, dir); err != nil {
		t.Errorf("cfg mock error %v", err)
	}
}
