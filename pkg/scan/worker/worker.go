// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"slices"
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

// Worker represents a job worker.
// blkSz unit is byte.
type Worker struct {
	scanner *yr.Scanner
	matches yr.MatchRules
	buff    []byte
	acts    *act.Cfg
	blkSz   int
	exp     time.Time
	expFn   func(time.Time, string) (bool, *unix.Stat_t)
	err     error
	offset  int
}

// New returns a worker from given cfg, rules and max file age.
func New(cfg *base.Cfg, rules *yr.Rules) (*Worker, error) {
	scanner, err := yr.NewScanner(rules)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", sig.ErrYrScanner, err)
	}

	blkSz := int(float64(cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2))

	worker := &Worker{
		scanner: scanner.SetFlags(yr.ScanFlagsFastMode),
		buff:    make([]byte, blkSz),
		acts:    cfg.Acts,
		blkSz:   blkSz,
		exp:     time.Now().AddDate(0, 0, -cfg.Scans.MaxAge),
		expFn:   noop,
	}

	if cfg.Scans.MaxAge != 0 {
		worker.expFn = fsys.IsExp
	}

	worker.scanner.SetCallback(&worker.matches)

	return worker, nil
}

// Work receives given queued paths to scan.
func (worker *Worker) Work(ctx context.Context, state *state.Job, queue chan string) {
	defer state.WGrp.Done()

	for path := range queue {
		select {
		case <-ctx.Done():
			return

		default:
			worker.Scan(path, state)
		}
	}
}

// Scan scans given file path and job state. Results and errs to job state.
func (worker *Worker) Scan(path string, result *state.Job) {
	exp, stat := worker.expFn(worker.exp, path)
	if exp {
		return
	}

	// Attempt to open the file even if stat failed.
	file, err := os.Open(path)
	if err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrFileRead, err, path))
		return
	}
	defer file.Close()

out:
	for {
		worker.offset, worker.err = file.Read(worker.buff)

		switch {
		case worker.offset > 0:
			if err := worker.scanner.ScanMem(worker.buff[:worker.offset]); err != nil {
				result.AddErr(fmt.Errorf("%w, %v, %v", ErrYrScan, err, path))
				return
			}

		case worker.offset == 0:
			break out
		}
	}

	if worker.err != nil && !errors.Is(worker.err, io.EOF) {
		result.AddErr(fmt.Errorf("%w, %v, %v", ErrFileRead, worker.err, path))
	}

	if len(worker.matches) == 0 {
		return
	}

	// Reuse from beginning of fn, otherwise start from scratch.
	if stat == nil {
		stat = &unix.Stat_t{}

		if err := unix.Stat(path, stat); err != nil {
			result.AddErr(fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, path))
		}
	}

	matches := MatchesToStr(worker.matches)

	// Reset per file.
	worker.matches = worker.matches[:0]

	result.Hits <- &state.Hit{
		Path: path,

		Meta: hit.NewMeta(
			fsys.NewAttr(stat),
			matches,
			worker.acts.NewVerbs(path, matches...)...,
		),
	}
}

// MatchesToStr returns a slice string from given yr.MatcheRules.
func MatchesToStr(matches yr.MatchRules) []string {
	rules := []string{}

	for _, rule := range matches {
		rules = append(rules, rule.Rule)
	}

	slices.Sort(rules)

	return slices.Compact(rules)
}

func noop(_ time.Time, _ string) (bool, *unix.Stat_t) {
	return false, nil
}

// Mock mocks a worker.
func Mock(env *env.Env) (*Worker, error) {
	if err := sig.Mock(env); err != nil {
		return nil, err
	}

	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		return nil, err
	}

	worker, err := New(env.Cfg, rules)
	if err != nil {
		return nil, err
	}

	return worker, nil
}
