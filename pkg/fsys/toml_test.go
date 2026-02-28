// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTOML(t *testing.T) {
	var (
		path = filepath.Join(t.TempDir(), t.Name())

		toml = `[example]
  key = "value"`
	)

	if err := os.WriteFile(path, []byte(toml), 0600); err != nil {
		t.Fatalf("unable to write to file: %v", err)
	}

	var cfg struct {
		Example struct {
			Key string `toml:"key"`
		} `toml:"example"`
	}

	if err := ReadTOML(path, &cfg); err != nil {
		t.Fatalf("read toml error: %v", err)
	}

	if cfg.Example.Key != "value" {
		t.Errorf("unexpected read toml result: got %v, want %v", cfg.Example.Key, "value")
	}
}

func TestInstallTOML(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if err := InstallTOML(input, struct{}{}); err != nil {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLExists(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	file, err := os.Create(input)
	if err != nil {
		t.Fatalf("toml write error: %v", err)
	}
	defer file.Close()

	if err := InstallTOML(input, &struct{}{}); !errors.Is(err, fs.ErrExist) {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLDisabled(t *testing.T) {
	var (
		path     = filepath.Join(t.TempDir(), t.Name()+".toml")
		disabled = strings.Replace(path, ".toml", ".disabled", 1)
	)

	file, err := os.Create(disabled)
	if err != nil {
		t.Fatalf("disabled toml write error: %v", err)
	}
	defer file.Close()

	if err := InstallTOML(path, struct{}{}); err != nil {
		t.Errorf("install toml error: %v", err)
	}
}

func TestInstallTOMLErrs(t *testing.T) {
	input := "/dev/null/" + t.Name()

	if err := InstallTOML(input, struct{}{}); !errors.Is(err, ErrTOMLRead) {
		t.Errorf("unexpected install toml error: %v", err)
	}
}

func TestWriteTOML(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())

	if err := WriteTOML(input, struct{}{}); err != nil {
		t.Errorf("write toml error: %v", err)
	}
}

func TestWriteTOMLErrs(t *testing.T) {
	input := "//"

	if err := WriteTOML(input, struct{}{}); !errors.Is(err, ErrFileOpen) {
		t.Errorf("unexpected write toml error: %v", err)
	}
}
