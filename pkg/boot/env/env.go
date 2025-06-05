// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package env

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/cmd"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

// Env represents env, used as a registry.
// State stores cmd state.
type Env struct {
	Opts  *Opts
	Plat  plat.Plat
	Paths *path.Paths
	Cfg   *base.Cfg
	Db    *bbolt.DB
	State *cmd.State
	Ver   string
}

// Opts represents cli opts.
// --no-alerts
// --no-ticker
// --unattended
type Opts struct {
	NoAlerts   bool
	NoTicker   bool
	Unattended bool
}

// ver stores the version. Embedded by ldflag during build.
var ver string

// Load discovers directory paths and loads the env.
func Load(state *cmd.State) (*Env, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrSelfBin, err)
	}

	// FUNCTIONAL PATHS
	var (
		cwd, bin = filepath.Split(self)
		conf     = filepath.Join(cwd, "cfg")
		sigs     = filepath.Join(cwd, "sigs")
		tmp      = filepath.Join(cwd, "tmp")
	)

	env := &Env{
		Opts: &Opts{},

		Paths: &path.Paths{
			Install: &path.Install{
				Dir:  cwd,
				Path: self,
				Bin:  bin,
				Tmp:  tmp,
			},

			Cfg: &path.Cfg{
				Dir:     conf,
				Base:    filepath.Join(conf, "cfg.toml"),
				Secrets: filepath.Join(conf, "secrets.toml"),
				Acts:    filepath.Join(conf, "actions.toml"),
			},

			Plat: &path.Plat{
				Dir: filepath.Join(conf, "plat"),
			},

			Sigs: &path.Sigs{
				Dir: sigs,
				Src: filepath.Join(sigs, "src"),
				Idx: filepath.Join(sigs, "index.yr"),
				Yrc: filepath.Join(sigs, "index.yrc"),
				Tmp: filepath.Join(tmp, "sigs"),
			},

			Alerts: &path.Alerts{
				Dir: filepath.Join(conf, "alerts"),
			},
		},

		State: state,

		Ver: ver,
	}

	if err := state.Lock(filepath.Join(cwd, bin+".lock")); err != nil {
		return nil, err
	}

	env.SetOpts()

	return env, nil
}

// SetOpts sets opts.
func (env *Env) SetOpts() {
	// Prevent flag conflicts.
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flag.BoolVar(&env.Opts.NoAlerts, "no-alerts", false, "disable alerts")
	flag.BoolVar(&env.Opts.NoTicker, "no-ticker", false, "disable progress ticker")
	flag.BoolVar(&env.Opts.Unattended, "unattended", false, "disable stdout")
	flag.Parse()
}

// Mock mocks for tests. Dir should be test's tmp dir. Name should be test name.
func Mock(name string, dir string) (*Env, error) {
	env := &Env{
		Opts: &Opts{},

		Plat: plat.Mock(acter.Mock(name, true)),

		Paths: &path.Paths{
			Install: &path.Install{
				Dir:  dir,
				Bin:  name,
				Path: filepath.Join(dir, name),
				Tmp:  filepath.Join(dir, "tmp"),
				Db:   filepath.Join(dir, name+".db"),
				Log:  filepath.Join(dir, name+".log"),
			},

			Cfg: &path.Cfg{
				Dir:     filepath.Join(dir, "cfg"),
				Base:    filepath.Join(dir, "cfg", "cfg.toml"),
				Secrets: filepath.Join(dir, "cfg", "secrets.toml"),
				Acts:    filepath.Join(dir, "cfg", "acts.toml"),
			},

			Plat: &path.Plat{
				Dir: filepath.Join(dir, "cfg", "plat"),
			},

			Sigs: &path.Sigs{
				Dir: dir,
				Src: filepath.Join(dir, "sigs", "src"),
				Idx: filepath.Join(dir, "sigs", "index.yr"),
				Yrc: filepath.Join(dir, "sigs", "index.yrc"),
				Tmp: filepath.Join(dir, "tmp", "sigs"),
			},

			Alerts: &path.Alerts{
				Dir: dir,
			},
		},

		State: cmd.NewState(),
	}

	base, err := base.Mock(env.Paths, dir)
	if err != nil {
		return nil, err
	}

	env.Cfg = base

	return env, nil
}
