// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"github.com/defended-net/malwatch/pkg/scan/state"
)

type mock struct {
	verb      string
	isEnabled bool
	Acted     bool
}

// Mock mocks an acter.
func Mock(verb string, enabled bool) Acter {
	return &mock{
		verb:      verb,
		isEnabled: enabled,
	}
}

// Load loads an exiler.
func (act *mock) Verb() string {
	return act.verb
}

// Load loads the quarantiner.
func (act *mock) Load() error {
	if !act.isEnabled {
		return ErrDisabled
	}

	return nil
}

// Run quarantines hits.
func (act *mock) Act(_ *state.Result) error {
	act.Acted = true

	return nil
}
