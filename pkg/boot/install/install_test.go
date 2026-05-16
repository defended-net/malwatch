// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestRunYesNo(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"Y": {
			input: "Y\n",
		},

		"y": {
			input: "y\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Fatalf("env mock error %v", err)
			}

			env.Paths.Cfg.Base += name

			rd, wr, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			defer func(file *os.File) {
				os.Stdin = file
			}(os.Stdin)

			os.Stdin = rd

			if _, err = wr.WriteString(test.input); err != nil {
				t.Fatal(err)
			}

			// lint
			_ = wr.Close()

			if got := Run(env); got != nil {
				t.Errorf("run error %v", got)
			}
		})
	}
}

func TestRunYesNoExit(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"N": {
			input: "N",
		},

		"n": {
			input: "n",
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error %v", err)
	}

	opt := os.Getenv("INPUT")

	if strings.ToLower(opt) == "n" {
		env.Paths.Cfg.Base += opt

		rd, wr, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}

		defer func(file *os.File) {
			os.Stdin = file
		}(os.Stdin)

		os.Stdin = rd

		if _, err = wr.WriteString(opt); err != nil {
			t.Fatal(err)
		}

		// lint
		_ = wr.Close()

		if err = Run(env); err != nil {
			t.Errorf("run err %v", err)
		}

		return
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
			cmd.Env = append(os.Environ(), "INPUT="+test.input)

			if e, ok := cmd.Run().(*exec.ExitError); ok && e.ExitCode() != 0 {
				t.Errorf("unexpected yes no result %v, want %v", e.ExitCode(), 0)
			}
		})
	}
}

func TestRunExists(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := Run(env); got != nil {
		t.Errorf("run err %v", got)
	}
}

func TestRunMkdirErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Cfg.Base += "-not-exist"

	path := filepath.Join(t.TempDir(), t.Name())

	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}

	env.Paths.Cfg.Dir = filepath.Join(path, "subdir")

	rdr, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	defer func(f *os.File) {
		os.Stdin = f
	}(os.Stdin)

	os.Stdin = rdr

	// lint
	_, _ = wr.WriteString("y\n")

	// lint
	_ = wr.Close()

	if got := Run(env); err == got {
		t.Error("unexpected run success")
	}
}

func TestRunStatErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error %v", err)
	}

	path := filepath.Join(t.TempDir(), t.Name())

	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}

	env.Paths.Cfg.Base = filepath.Join(path, "not-exist")

	if got := Run(env); got == nil {
		t.Error("unexpected run success")
	}
}

func TestSysd(t *testing.T) {
	if os.Getuid() != 0 {
		fmt.Println("install: systemd tests require root")
		return
	}

	input := t.TempDir()

	if err := Sysd(input, filepath.Join(input, t.Name())); err != nil {
		t.Errorf("sysd err %v", err)
	}
}

func TestSysdUnsupported(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("test requires non-root")
	}

	var (
		got  = Sysd(t.TempDir(), filepath.Join(t.TempDir(), "bin"))
		want = "install: systemd support require root"
	)

	if got == nil {
		t.Error("unexpected sysd success")
	}

	if got.Error() != want {
		t.Errorf("unexpected sysd err %v, want %v", got, want)
	}
}

func TestSysdMonitorSkip(t *testing.T) {
	if got := Sysd(t.TempDir(), "-monitor"); got != nil {
		t.Errorf("sysd err %v", got)
	}
}
