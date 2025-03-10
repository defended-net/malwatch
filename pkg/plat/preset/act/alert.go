// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/alert"
	"github.com/defended-net/malwatch/pkg/plat/preset/alert/json"
	"github.com/defended-net/malwatch/pkg/plat/preset/alert/pagerduty"
	"github.com/defended-net/malwatch/pkg/plat/preset/alert/smtp"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Alerter represents alerting.
type Alerter struct {
	verb    string
	senders []alert.Sender
}

// NewAlerter returns alerter for given env.
func NewAlerter(env *env.Env) *Alerter {
	return &Alerter{
		verb: VerbAlert,

		senders: []alert.Sender{
			json.New(env),
			pagerduty.New(env),
			smtp.New(env),
		},
	}
}

// Load loads given alerter.
func (acter *Alerter) Load() error {
	enabled := []alert.Sender{}

	for _, alerter := range acter.senders {
		err := fsys.InstallTOML(alerter.Cfg().Path(), alerter.Cfg())

		switch {
		case err == nil:
			continue

		case errors.Is(err, fs.ErrExist):
			if err := alerter.Load(); err != nil {
				return err
			}

			enabled = append(enabled, alerter)

		default:
			return fmt.Errorf("%w, %v, %v", ErrCfgLoad, err, alerter.Cfg().Path())
		}
	}

	if len(enabled) == 0 {
		return ErrDisabled
	}

	acter.senders = enabled

	return nil
}

// Act sends alerts for given result.
func (acter *Alerter) Act(result *state.Result) error {
	for _, alerter := range acter.senders {
		if err := alerter.Alert(result); err != nil {
			result.AddErr(err)
		}
	}

	return nil
}

// Verb returns a given alerter verb.
func (acter *Alerter) Verb() string {
	return acter.verb
}
