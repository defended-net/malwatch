// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package alert

import (
	"os"

	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Sender represents sender.
type Sender interface {
	Load(*os.Root) error
	Cfg() plat.Cfg
	Alert(*state.Result) error
}
