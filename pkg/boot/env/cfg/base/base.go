// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package base

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/db"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/logger"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/scan"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents base cfg. It also references secondary cfgs such as secrets, actions, etc.
type Cfg struct {
	path       string
	Identifier string `env:"IDENTIFIER"`
	Cores      int    `env:"CORES"`
	Threads    int    `env:"THREADS"`
	Scans      *scan.Cfg
	Secrets    *secret.Cfg
	Acts       *act.Cfg
	Database   *db.Cfg
	Log        *logger.Cfg
}

// New returns a new base cfg from given paths.
func New(paths *path.Paths) *Cfg {
	return &Cfg{
		path: paths.Cfg.Base,

		Scans: &scan.Cfg{
			Monitor: &scan.Monitor{},
		},

		Database: &db.Cfg{
			Dir: paths.Install.Db,
		},

		Log: &logger.Cfg{
			Dir: paths.Install.Log,
		},
	}
}

// Load reads the cfg from toml path.
func (cfg *Cfg) Load() error {
	if err := fsys.ReadTOML(cfg.Path(), cfg); err != nil {
		return err
	}

	return cfg.Validate()
}

// Validate validates cfg var values beyond toml concrete types.
func (cfg *Cfg) Validate() error {
	if cfg.Identifier == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("%w, %v", ErrHostnameLookup, err)
		}

		cfg.Identifier = hostname
	}

	if cfg.Threads == 0 {
		cfg.Threads = runtime.NumCPU()
	}

	var targets []*regexp.Regexp

	for _, target := range cfg.Scans.Targets {
		re := regexp.MustCompile(target)

		targets = append(targets, re)
	}

	re.SetTargets(targets...)

	return nil
}

// Mock mocks a cfg.
func Mock(paths *path.Paths, dir string) (*Cfg, error) {
	cfg := &Cfg{
		path: filepath.Join(dir, "cfg", "cfg.toml"),

		Cores:   1,
		Threads: 1,

		Database: &db.Cfg{
			Dir: filepath.Join(dir, "db"),
		},

		Scans: &scan.Cfg{
			Targets: []string{
				`^/(?P<target>target)/?(.*)`,
			},

			BlkSz: 65536,

			Paths: []string{
				dir,
			},

			Monitor: &scan.Monitor{
				Timeout: 5,
			},
		},

		Acts: act.Mock(paths.Cfg.Acts),
		Log:  logger.Mock(filepath.Dir(paths.Install.Log)),
	}

	secrets, err := secret.Mock(paths.Cfg.Secrets)
	if err != nil {
		return nil, err
	}

	cfg.Secrets = secrets

	dirs := []string{
		cfg.Log.Dir,
		cfg.Database.Dir,

		paths.Cfg.Dir,
		paths.Plat.Dir,
		paths.Alerts.Dir,
		paths.Sigs.Src,
		paths.Sigs.Tmp,
		paths.Install.Tmp,
	}

	files := []string{
		paths.Cfg.Base,
		paths.Cfg.Secrets,
		paths.Cfg.Acts,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, err
		}
	}

	for _, path := range files {
		if _, err := os.Create(path); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}
