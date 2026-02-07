// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"github.com/defended-net/malwatch/third_party/yr"
)

// Scanner represents yara.
type Scanner struct {
	Val  *yr.Scanner
	Rev  uint64
	GcFn func()
}

// Gc does cleanup.
func (scanner *Scanner) Gc() {
	if scanner.Val != nil {
		scanner.Val.Destroy()
		scanner.Val = nil
	}

	if scanner.GcFn != nil {
		scanner.GcFn()
		scanner.GcFn = nil
	}
}
