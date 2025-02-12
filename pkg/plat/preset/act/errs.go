// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import "errors"

var (
	// ErrDisabled means disabled.
	ErrDisabled = errors.New("act: disabled")

	// ErrCfgLoad means cfg load error.
	ErrCfgLoad = errors.New("act: cfg load error")
)

// ALERTS
var (
	// ErrAlerterLoad means alerter load error.
	ErrAlerterLoad = errors.New("act: alerter load error")

	// ErrAlertSend means alerter load error.
	ErrAlertSend = errors.New("act: alert send error")
)

// QUARANTINES
var (
	// ErrQuarantineNoDir means no quarantine dir.
	ErrQuarantineNoDir = errors.New("act: no quarantine dir configured in cfg/actions.toml")
)

// CLEANS
var (
	// ErrCleanNoExpr means no expr for match.
	ErrCleanNoExpr = errors.New("act: no clean expressions for match")

	// ErrCleanFailed means still hits after clean.
	ErrCleanFailed = errors.New("act: clean failed")
)

// EXILES
var (
	// ErrExileNoRegion means no region configured.
	ErrExileNoRegion = errors.New("act: no region configured for exile")

	// ErrExileUpload means upload error.
	ErrExileUpload = errors.New("act: exile upload error")
)
