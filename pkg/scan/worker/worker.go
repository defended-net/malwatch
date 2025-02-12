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
// isExpired is optimisation for if max age check is disabled.
// blkSz unit is byte.
// expiry unit is day.
type Worker struct {
	scanner   *yr.Scanner
	matches   yr.MatchRules
	buff      []byte
	act       *act.Cfg
	blkSz     int
	expiry    time.Time
	isExpired func(time.Time, string) (*unix.Stat_t, bool)
}

// New returns a worker from given cfg, rules and max file age.
func New(cfg *base.Cfg, rules *yr.Rules, maxAge int) (*Worker, error) {
	scanner, err := yr.NewScanner(rules)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", sig.ErrYrScanner, err)
	}

	scanner.SetFlags(yr.ScanFlagsFastMode)

	blkSz := int(float64(cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2))

	worker := &Worker{
		scanner:   scanner,
		matches:   yr.MatchRules{},
		buff:      make([]byte, blkSz),
		act:       cfg.Acts,
		blkSz:     blkSz,
		expiry:    time.Now().AddDate(0, 0, -maxAge),
		isExpired: func(_ time.Time, _ string) (*unix.Stat_t, bool) { return nil, false },
	}

	if maxAge != 0 {
		worker.isExpired = fsys.IsExpired
	}

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

// Scan scans a given file path to a channel of data blocks.
// Errs are stored as mutex based concurrent data structure.
func (worker *Worker) Scan(path string, result *state.Job) {
	// Returning a bool is quicker compared to nil check for worker.isExpired.
	stat, expired := worker.isExpired(worker.expiry, path)
	if expired {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		result.AddErr(fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, path))
		return
	}
	defer file.Close()

	worker.scanner.SetCallback(&worker.matches)

	rules := []string{}

	for {
		offset, err := file.Read(worker.buff)

		if offset > 0 {
			if err := worker.scanner.ScanMem(worker.buff[:offset]); err != nil {
				result.AddErr(fmt.Errorf("%w, %v, %v", ErrYrScan, err, path))
				return
			}

			rules = append(rules, MatchesToString(worker.matches)...)

			worker.buff = worker.buff[:worker.blkSz]
			worker.matches = nil

			// We still want to scan whatever we can get, keep going.
			continue
		}

		// Deal with error after examining everything we can possibly get.
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// Something bad happened, we have done what we can so bail now.
				result.AddErr(fmt.Errorf("%w, %v, %v", ErrFileRead, err, path))
			}

			break
		}
	}

	if len(rules) == 0 {
		return
	}

	// Reuse from beginning of fn, otherwise start from scratch.
	if stat == nil {
		stat = &unix.Stat_t{}

		if err := unix.Stat(path, stat); err != nil {
			result.AddErr(fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, path))
		}
	}

	result.Hits <- &state.Hit{
		Path: path,
		Meta: hit.NewMeta(fsys.NewAttr(stat), rules, worker.act.NewVerbs(path, rules...)...),
	}
}

// MatchesToString returns a slice string from given yr.MatcheRules.
func MatchesToString(matches yr.MatchRules) []string {
	rules := []string{}

	for _, rule := range matches {
		rules = append(rules, rule.Rule)
	}

	slices.Sort(rules)

	return slices.Compact(rules)
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

	worker, err := New(env.Cfg, rules, 0)
	if err != nil {
		return nil, err
	}

	return worker, nil

}
