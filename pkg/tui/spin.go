// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Spinner represents a progress spinner.
type Spinner struct {
	msg      string
	interval time.Duration
	running  atomic.Bool
	once     sync.Once
	done     sync.WaitGroup
}

// NewSpinner returns a spinner from given interval and message.
func NewSpinner(interval time.Duration, msg string) *Spinner {
	return &Spinner{
		interval: interval,
		msg:      msg,
	}
}

// Start starts a spinner.
func (spinner *Spinner) Start() {
	if spinner == nil {
		return
	}

	spinner.running.Store(true)

	spinner.done.Add(1)
	defer spinner.done.Done()

	var (
		ticker = time.NewTicker(spinner.interval)
		idx    = 0
		tick   = []string{"-", "\\", "|", "/"}
	)

	defer ticker.Stop()

	for spinner.running.Load() {
		<-ticker.C

		fmt.Printf("\r%v [%v]", spinner.msg, tick[idx])
		idx = (idx + 1) % len(tick)
	}

	fmt.Printf("\r\033[2K")
}

// Stop stops a spinner.
func (spinner *Spinner) Stop() {
	if spinner == nil {
		return
	}

	spinner.once.Do(func() {
		spinner.running.Store(false)
		spinner.done.Wait()
	})
}
