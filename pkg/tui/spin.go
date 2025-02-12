// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"fmt"
	"time"
)

// Spinner represents a progress spinner.
type Spinner struct {
	tick     *time.Ticker
	interval time.Duration
	msg      string
	stop     chan (struct{})
}

// NewSpinner returns a spinner from given interval and message.
func NewSpinner(interval time.Duration, msg string) *Spinner {
	return &Spinner{
		tick:     time.NewTicker(interval),
		interval: interval,
		msg:      msg,
		stop:     make(chan struct{}, 1),
	}
}

// Start starts a spinner.
func (spinner *Spinner) Start() {
	if spinner == nil {
		return
	}

	defer close(spinner.stop)
	defer spinner.Stop()

	tick := []string{
		"-",
		"\\",
		"|",
		"/",
	}

	idx := 0

	for {
		select {
		case <-spinner.stop:
			return

		case <-spinner.tick.C:
			fmt.Printf("\r%v [%v]", spinner.msg, tick[idx])

			idx++

			if idx > 3 {
				idx = 0
			}
		}
	}
}

// Stop stops a spinner.
func (spinner *Spinner) Stop() {
	if spinner == nil {
		return
	}

	spinner.tick.Stop()
	spinner.stop <- struct{}{}

	// Clear.
	fmt.Printf("\r")
	fmt.Printf("\033[2K")
}
