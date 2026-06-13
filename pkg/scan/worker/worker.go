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
	exp     time.Time
	expFn   func(time.Time, int) (bool, *unix.Stat_t)
}

// matches represents matched rules.
type matches []string

// New returns a worker from given cfg, rules and max file age.
func New(cfg *base.Cfg) (*Worker, error) {
	// #nosec G404 -- non crypto jitter.
	blkSz := int(float64(cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2))

	worker := &Worker{
		buff:  make([]byte, blkSz),
		acts:  cfg.Acts,
		exp:   time.Now().AddDate(0, 0, -cfg.Scans.MaxAge),
		expFn: noop,
	}

	if cfg.Scans.MaxAge != 0 {
		worker.expFn = fsys.IsExp
	}

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

	exp, stat := worker.expFn(worker.exp, fd)
	if exp {
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

	// Reuse from beginning of fn, otherwise start from scratch.
	if stat == nil {
		stat = &unix.Stat_t{}

		if err := unix.Fstat(fd, stat); err != nil {
			result.AddErr(fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, path))
		}
	}

	matches := slices.Clone(worker.matches)
	slices.Sort(matches)

	result.Hits <- &state.Hit{
		Path: path,

		Meta: hit.NewMeta(
			fsys.NewAttr(stat),
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

func noop(_ time.Time, _ int) (bool, *unix.Stat_t) {
	return false, nil
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
