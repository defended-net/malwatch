// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"errors"
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

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
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
			wr.Close()

			if err = Run(env); err != nil {
				t.Errorf("run error: %v", err)
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
		t.Fatalf("env mock error: %v", err)
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
		wr.Close()

		if err = Run(env); err != nil {
			t.Errorf("yesno exit run error: %v", err)
		}

		return
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
			cmd.Env = append(os.Environ(), "INPUT="+test.input)

			if e, ok := cmd.Run().(*exec.ExitError); ok && e.ExitCode() != 0 {
				t.Errorf("unexpected yes no exit code: %v, want %v", e.ExitCode(), 0)
			}
		})
	}
}

func TestRunExists(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err = Run(env); err != nil {
		t.Errorf("run error: %v", err)
	}
}

func TestSysd(t *testing.T) {
	if os.Getuid() != 0 {
		fmt.Println("install: systemd tests require root")
		return
	}

	input := t.TempDir()

	if err := Sysd(input, filepath.Join(input, t.Name())); err != nil {
		t.Errorf("systemd error: %v", err)
	}
}

func TestSysdUnsupported(t *testing.T) {
	if os.Getuid() != 0 {
		fmt.Println("install: systemd tests require root")
		return
	}

	input := filepath.Join(t.TempDir(), "not-exist")

	if err := Sysd(input, filepath.Join(input, t.Name())); !errors.Is(err, ErrSysdMissing) {
		t.Errorf("systemd error: %v", err)
	}
}
