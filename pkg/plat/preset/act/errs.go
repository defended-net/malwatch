// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import "errors"

var (
	// ErrCfgLoad means cfg load error.
	ErrCfgLoad = errors.New("act: cfg load error")
)

// ALERTS
var (
	// ErrAlerterLoad means alerter load err.
	ErrAlerterLoad = errors.New("act: alerter load error")

	// ErrAlertSend means alert send err.
	ErrAlertSend = errors.New("act: alert send error")
)

// QUARANTINES
var (
	// ErrQuarantineNoDir means no quarantine dir.
	ErrQuarantineNoDir = errors.New("act: no quarantine dir configured in cfg/actions.toml")

	// ErrQuarantineMv means hit mv err.
	ErrQuarantineMv = errors.New("act: move hit error")
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

	// ErrExileUpload means upload err.
	ErrExileUpload = errors.New("act: exile upload error")

	// ErrExileDelErr means del err.
	ErrExileDelErr = errors.New("act: exile del error")
)
