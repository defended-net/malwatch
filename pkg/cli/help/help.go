// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package help

import (
	"errors"
)

// Help represents cli usage.
type Help struct {
	Arg  string
	Desc error
}

// Helps represents helps.
type Helps []Help

// Logo is the logo.
const Logo = "              __            __      __\n  __ _  ___ _/ /    _____ _/ /_____/ /\n /    \\/ _ `/ / |/|/ / _ `/ __/ __/ _ \\\n/_/_/_/\\_,_/_/|__,__/\\_,_/\\__/\\__/_//_/\n"

var (
	// ErrScan means scan help.
	ErrScan = errors.New("malwatch scan [PATH]")

	// ErrHistory means history help.
	ErrHistory = errors.New("malwatch history {get | del} [TARGET | PATH]")

	// ErrActs means acts help.
	ErrActs = errors.New("malwatch actions {get | set | del} PATH SIGNATURE [ACTION...]")

	// ErrQuarantine means quarantine help.
	ErrQuarantine = errors.New("malwatch quarantine PATH")

	// ErrExile means exile help.
	ErrExile = errors.New("malwatch exile PATH")

	// ErrRestore means restore help.
	ErrRestore = errors.New("malwatch restore {PATH | SCAN_ID}")

	// ErrSigs means sig help.
	ErrSigs = errors.New("malwatch signatures {update | refresh}")

	// ErrInfo means info help.
	ErrInfo = errors.New("malwatch info")

	// ErrInstall means install help.
	ErrInstall = errors.New("malwatch install")
)

// malwatch-monitor
var (
	// ErrMonitor means monitor help.
	ErrMonitor = errors.New("malwatch-monitor start")
)
