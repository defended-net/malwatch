// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package env

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/cmd"
)

func TestLoad(t *testing.T) {
	os.Args = []string{
		"malwatch",
	}

	flag.Parse()

	path, err := os.Executable()
	if err != nil {
		t.Fatalf("exec path lookup error: %v", err)
	}

	var (
		cwd, _ = filepath.Split(path)
		dir    = filepath.Join(cwd, "cfg")

		files = []string{
			filepath.Join(dir, "actions.toml"),
			filepath.Join(dir, "secrets.toml"),
			filepath.Join(dir, "clean.toml"),
		}
	)

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	for _, path := range files {
		if _, err := os.Create(path); err != nil {
			t.Fatalf("file create error: %v", err)
		}
	}

	if _, err = Load(&cmd.State{}); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestSetOps(t *testing.T) {
	env, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	layout := map[string]*bool{
		"-unattended": &env.Opts.Unattended,
		"-no-alerts":  &env.Opts.NoAlerts,
		"-no-ticker":  &env.Opts.NoTicker,
	}

	for arg, opt := range layout {
		t.Run(arg, func(t *testing.T) {
			os.Args = []string{
				"malwatch",
				arg,
				"version",
			}

			env.SetOpts()

			if !*opt {
				t.Errorf("unexpected set opt result %v, want %v", *opt, true)
			}
		})
	}
}
