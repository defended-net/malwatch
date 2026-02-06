// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

type monitor struct {
	rev  atomic.Uint64
	path string
	cmp  *unix.Stat_t
	prev *unix.Stat_t
}

// Monitor monitors for rule changes.
func Monitor(env *env.Env) error {
	ctx, cancel := context.WithCancel(context.Background())
	env.State.AddCancel(cancel)

	var (
		monitor = &monitor{
			path: env.Paths.Sigs.Yrc,
			prev: &unix.Stat_t{},
			cmp:  &unix.Stat_t{},
		}

		timer = time.NewTimer(time.Second)
	)

	if err := unix.Stat(monitor.path, monitor.prev); err != nil {
		return fmt.Errorf("%w, %v, %v, %v, %v", ErrYrcStat, "path", monitor.path, "err", err)
	}

	if err := Set(monitor.path, 0); err != nil {
		return fmt.Errorf("%w, %v, %v, %v, %v", ErrYrcSet, "path", monitor.path, "err", err)
	}

	go func() {
		defer func() {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-timer.C:
				monitor.tick()
				timer.Reset(time.Second)
			}
		}
	}()

	return nil
}

// tick does one monitoring cycle.
func (monitor *monitor) tick() {
	if err := unix.Stat(monitor.path, monitor.cmp); err != nil {
		slog.Info(ErrYrcStat.Error(), "rev", monitor.rev.Load(), "path", monitor.path, "err", err)
		return
	}

	if monitor.cmp.Mtim == monitor.prev.Mtim {
		return
	}

	// debounce
	time.Sleep(300 * time.Millisecond)

	*monitor.prev = *monitor.cmp
	monitor.rev.Add(1)

	if err := Set(monitor.path, monitor.rev.Load()); err != nil {
		slog.Info(ErrYrcSet.Error(), "rev", monitor.rev.Load(), "path", monitor.path, "err", err)
		return
	}

	slog.Info("sig: refreshed", "rev", monitor.rev.Load(), "path", monitor.path)
}
