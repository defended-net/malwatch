// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Acter represents an acter.
type Acter interface {
	Load() error
	Verb() string
	Act(*state.Result) error
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
