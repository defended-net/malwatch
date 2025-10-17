// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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

// Run performs initial install.
func Run(env *env.Env) error {
	// cfg dir exists, abort.
	if _, err := os.Stat(env.Paths.Cfg.Base); err == nil {
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
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	env.Cfg = base.New(env.Paths)
	env.Cfg.Scans = scan.New()

	// Before proceeding, let's write. Referenced fields (acts, etc)
	// can then be excluded to ensure a lean file.
	if err := fsys.WriteTOML(env.Paths.Cfg.Base, env.Cfg); err != nil {
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
		if err := fsys.WriteTOML(cfg.Path(), cfg); err != nil {
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

	dst := filepath.Join(sysdDir, "malwatch-monitor.service")

	if _, err := os.Stat(sysdDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w, %v", ErrSysdMissing, sysdDir)
		}

		return err
	}

	// runlevel 3, 4, 5
	cfg := `[Unit]
Description=malwatch-monitor

[Service]
ExecStart=` + binPath + ` start

[Install]
WantedBy=multi-user.target
`

	return os.WriteFile(dst, []byte(cfg), 0600)
}
