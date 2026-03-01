// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// InstallTOML installs toml file. Checks for existing .toml file.
// If not exist, then write file but with ext .disabled.
func InstallTOML(path string, cfg any) error {
	// .toml exists, abort.
	if _, err := toml.DecodeFile(path, cfg); err == nil {
		// Should not be logged.
		return fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	disabled := strings.TrimSuffix(path, ".toml") + ".disabled"

	// .disabled exists, abort.
	if _, err := os.Stat(disabled); err == nil {
		return nil
	}

	return WriteTOML(disabled, cfg)
}

// ReadTOML reads toml file for given cfg.
func ReadTOML(path string, cfg any) error {
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	return nil
}

// WriteTOML (over)writes toml file with given cfg.
func WriteTOML(path string, cfg any) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, path)
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLWrite, err, path)
	}

	return nil
}
