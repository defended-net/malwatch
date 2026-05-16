// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// InstallTOML installs toml file. Checks for existing .toml file.
// If not exist, then write file but with ext .disabled.
func InstallTOML(root *os.Root, name string, cfg any) error {
	if root == nil {
		return ErrPathRoot
	}

	name, err := RootName(root, name)
	if err != nil {
		return err
	}

	path := filepath.Join(root.Name(), name)

	// .toml exists, abort.
	file, err := root.Open(name)
	if err == nil {
		defer Close(file)

		if _, err := toml.NewDecoder(file).Decode(cfg); err != nil {
			return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
		}

		// Should not be logged.
		return fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	disabled := strings.TrimSuffix(name, ".toml") + ".disabled"

	if !filepath.IsLocal(disabled) {
		return fmt.Errorf("%w, %v", ErrPathLocal, filepath.Join(root.Name(), disabled))
	}

	out, err := root.OpenFile(disabled, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, filepath.Join(root.Name(), disabled))
	}
	defer Close(out)

	if err := toml.NewEncoder(out).Encode(cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLWrite, err, filepath.Join(root.Name(), disabled))
	}

	return nil
}

// ReadTOML reads toml file for given cfg through root.
func ReadTOML(root *os.Root, name string, cfg any) error {
	if root == nil {
		return ErrPathRoot
	}

	name, err := RootName(root, name)
	if err != nil {
		return err
	}

	path := filepath.Join(root.Name(), name)

	file, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}
	defer Close(file)

	if _, err := toml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLRead, err, path)
	}

	return nil
}

// WriteTOML overwrites toml file with given cfg.
func WriteTOML(root *os.Root, name string, cfg any) error {
	if root == nil {
		return ErrPathRoot
	}

	name, err := RootName(root, name)
	if err != nil {
		return err
	}

	path := filepath.Join(root.Name(), name)

	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrFileOpen, err, path)
	}
	defer Close(file)

	if err := toml.NewEncoder(file).Encode(cfg); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrTOMLWrite, err, path)
	}

	return err
}
