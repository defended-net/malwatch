// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/yr"
)

// Worker represents a worker.
// blkSz unit is byte.
type Worker struct {
	scanner atomic.Pointer[state.Scanner]
	matches matches
	acts    *act.Cfg
	buff    []byte
	stat    *unix.Stat_t
	maxAge  int
	exp     time.Time
}

// matches represents matched rules.
type matches []string

// New returns a worker from given cfg, rules and max file age.
func New(cfg *base.Cfg) (*Worker, error) {
	var (
		// #nosec G404 -- non crypto jitter.
		blkSz = int(float64(cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2))

		worker = &Worker{
			acts:   cfg.Acts,
			buff:   make([]byte, blkSz),
			stat:   &unix.Stat_t{},
			maxAge: cfg.Scans.MaxAge,
			exp:    time.Now().AddDate(0, 0, -cfg.Scans.MaxAge),
		}
	)

	if err := worker.Refresh(); err != nil {
		return nil, err
	}

	return worker, nil
}

// Work receives given queued paths to scan.
func (worker *Worker) Work(ctx context.Context, state *state.Job, queue <-chan string) {
	defer state.WGrp.Done()

	done := ctx.Done()

	for {
		select {
		case <-done:
			return

		case path, ok := <-queue:
			if !ok {
				return
			}

			worker.Scan(path, state)
		}
	}
}

// Scan scans given file path and job state. Results and errs to job state.
func (worker *Worker) Scan(path string, result *state.Job) {
	worker.matches = worker.matches[:0]

	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
	if err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrFileRead, err, path))

		return
	}

	// nolint
	defer unix.Close(fd)

	if worker.maxAge != 0 && fsys.IsExp(worker.exp, fd, worker.stat) {
		return
	}

	var (
		scanner = worker.scanner.Load()
		offset  int
	)

out:
	for {
		offset, err = unix.Read(fd, worker.buff)

		switch {
		case offset > 0:
			if err := scanner.Val.ScanMem(worker.buff[:offset]); err != nil {
				result.AddErr(fmt.Errorf("%w, %v, %v", ErrYrScan, err, path))

				return
			}

			// condemn
			if len(worker.matches) > 3 {
				break out
			}

		default:
			break out
		}
	}

	if err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrFileRead, err, path))
	}

	if len(worker.matches) == 0 {
		return
	}

	matches := slices.Clone(worker.matches)
	slices.Sort(matches)

	if err := unix.Fstat(fd, worker.stat); err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, path))
	}

	result.Hits <- &state.Hit{
		Path: path,

		Meta: hit.NewMeta(
			fsys.NewAttr(worker.stat),
			matches,
			worker.acts.NewVerbs(path, matches...)...,
		),
	}
}

// Refresh updates the worker to use the latest sigs with a new scanner.
func (worker *Worker) Refresh() error {
	sigs, err := sig.Acquire()
	if err != nil {
		return err
	}

	scanner, err := yr.NewScanner(sigs.Rules)
	if err != nil {
		sigs.Release()
		return err
	}

	scanner.SetFlags(yr.ScanFlagsFastMode)
	scanner.SetCallback(&worker.matches)

	update := &state.Scanner{
		Val:  scanner,
		Rev:  sigs.Rev,
		GcFn: sigs.Release,
	}

	if old := worker.scanner.Swap(update); old != nil {
		old.Gc()
	}

	return nil
}

// RuleMatching implements yr.ScanCallback.
func (matches *matches) RuleMatching(_ *yr.ScanContext, rule *yr.Rule) (bool, error) {
	hit := rule.Identifier()

	if slices.Contains(*matches, hit) {
		return false, nil
	}

	*matches = append(*matches, hit)

	return len(*matches) > 3, nil
}

// Mock mocks a worker.
func Mock(env *env.Env) (*Worker, error) {
	if err := sig.Mock(env, true); err != nil {
		return nil, err
	}

	worker, err := New(env.Cfg)
	if err != nil {
		return nil, err
	}

	return worker, nil
}
