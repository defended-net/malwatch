// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"errors"
	"os"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Acter represents an acter.
type Acter interface {
	Load(*os.Root) error
	Verb() string
	Act(*state.Result) error
}

// Load loads given acters filtering on enabled.
func Load(root *os.Root, acters []Acter) ([]Acter, error) {
	enabled := []Acter{}

	for _, acter := range acters {
		err := acter.Load(root)

		switch {
		case errors.Is(err, ErrDisabled):
			continue

		case err != nil:
			return nil, err
		}

		enabled = append(enabled, acter)
	}

	return enabled, nil
}

// Get returns a verb's acter.
func Get(acts []Acter, verb string) (Acter, error) {
	for _, acter := range acts {
		if acter.Verb() == verb {
			return acter, nil
		}
	}

	return nil, ErrVerbUnknown
}

// Do gets the actfn from given acters and verb for invocation.
func Do(acters []Acter, verb string, result *state.Result) error {
	acter, err := Get(acters, verb)
	if err != nil {
		return err
	}

	return acter.Act(result)
}
