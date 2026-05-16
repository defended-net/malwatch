// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/scan"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/cmd"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/pkg/tui"
)

var tmpl = template.Must(
	template.New("unit").Parse(
		`[Unit]
Description=malwatch-monitor

[Service]
ExecStart={{.BinPath}} start

[Install]
WantedBy=multi-user.target
`))

// Run performs initial install.
func Run(env *env.Env) error {
	baseName, err := fsys.RootName(env.Paths.Install.Root, env.Paths.Cfg.Base)
	if err != nil {
		return err
	}

	// cfg dir exists, abort.
	if _, err := env.Paths.Install.Root.Stat(baseName); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if len(os.Args) == 1 || os.Args[1] != "install" {
		if ok := tui.YesNo("initial run, proceed to install?", os.Stdin); !ok {
			cmd.Exit(env.State, nil)
		}
	}

	// First create needed dirs.
	for _, dir := range []string{
		env.Paths.Cfg.Dir,
		env.Paths.Alerts.Dir,
		env.Paths.Sigs.Dir,
		env.Paths.Plat.Dir,
		env.Paths.Install.Tmp,
	} {
		name, err := fsys.RootName(env.Paths.Install.Root, dir)
		if err != nil {
			return err
		}

		if err := env.Paths.Install.Root.MkdirAll(name, 0700); err != nil {
			return err
		}
	}

	env.Cfg = base.New(env.Paths)
	env.Cfg.Scans = scan.New()

	// Before proceeding, let's write. Referenced fields (acts, etc)
	// can then be excluded to ensure a lean file.
	if err := fsys.WriteTOML(env.Paths.Install.Root, baseName, env.Cfg); err != nil {
		return err
	}

	env.Cfg.Acts = act.New(env.Paths.Cfg.Acts)
	env.Cfg.Secrets = secret.New(env.Paths.Cfg.Secrets)

	env.Cfg.Secrets.Submit.Endpoint = "https://api.defended.net/malwatch/submit"

	env.Cfg.Acts.Default = []string{"alert"}
	env.Cfg.Acts.Clean = act.Clean{
		"php_base64_inject": {
			"s/<?.*eval\\(base64_decode\\(.*?>//",
			"s/<?php.*eval\\(base64_decode\\(.*?>//",
			"s/eval\\(base64_decode\\([^;]*;//",
		},

		"php_gzbase64_inject": {
			"s/<?.*eval\\(gzinflate\\(base64_decode\\(.*?>//",
			"s/<?php.*eval\\(gzinflate\\(base64_decode\\(.*?>//",
			"s/eval\\(gzinflate\\(base64_decode\\(.*\\);//",
		},
	}

	env.Cfg.Secrets.Git = []*secret.Repo{
		{
			URL: "https://github.com/defended-net/malwatch-signatures",
		},
	}

	for _, cfg := range []cfg.Cfg{
		env.Cfg.Acts,
		env.Cfg.Secrets,
	} {
		if err := fsys.WriteTOML(env.Paths.Install.Root, cfg.Path(), cfg); err != nil {
			return err
		}
	}

	return sig.Update(env)
}

// Sysd installs a systemd profile for malwatch-monitor.
// Lowercase 'd' https://en.wikipedia.org/wiki/Systemd
func Sysd(sysdDir string, binPath string) error {
	// Avoid setting up systemd during tests.
	if binPath == "-monitor" {
		return nil
	}

	if os.Getuid() != 0 {
		return fmt.Errorf("install: systemd support require root")
	}

	slog.Info("installing systemd unit")

	root, err := os.OpenRoot(sysdDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w, %v", ErrSysdMissing, sysdDir)
		}

		return err
	}
	defer fsys.Close(root)

	file, err := root.OpenFile("malwatch-monitor.service", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%w, %v", err, filepath.Join(sysdDir, "malwatch-monitor.service"))
	}
	defer fsys.Close(file)

	return tmpl.Execute(
		file,

		struct{ BinPath string }{
			BinPath: binPath,
		},
	)
}
