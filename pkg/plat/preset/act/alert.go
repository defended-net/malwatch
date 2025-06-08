// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
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
func (alerter *Alerter) Load() error {
	enabled := []alert.Sender{}

	for _, sender := range alerter.senders {
		err := fsys.InstallTOML(sender.Cfg().Path(), sender.Cfg())

		switch {
		case err == nil:
			continue

		case errors.Is(err, fs.ErrExist):
			err := sender.Load()

			switch {
			case errors.Is(err, acter.ErrDisabled):
				continue

			case err != nil:
				return err
			}

			enabled = append(enabled, sender)

		default:
			return fmt.Errorf("%w, %v, %v", ErrCfgLoad, err, sender.Cfg().Path())
		}
	}

	alerter.senders = enabled

	if len(enabled) == 0 {
		return acter.ErrDisabled
	}

	return nil
}

// Act sends alerts for given result.
func (alerter *Alerter) Act(result *state.Result) error {
	for _, sender := range alerter.senders {
		if err := sender.Alert(result); err != nil {
			result.AddErr(err)
		}
	}

	return nil
}

// Verb returns a given alerter verb.
func (alerter *Alerter) Verb() string {
	return alerter.verb
}
