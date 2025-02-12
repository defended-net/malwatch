// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import "errors"

var (
	// ErrNoAPIToken means no api token in cfg.
	ErrNoAPIToken = errors.New("pagerduty: unable to read api token")

	// ErrNoSeverity means no severity in cfg.
	ErrNoSeverity = errors.New("pagerduty: no severity configured")
)
