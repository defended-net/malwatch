// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwtodd/Go.Sed/sed"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/yr"
)

// Cleaner represents cleaning.
type Cleaner struct {
	verb    string
	dir     string
	blkSz   int
	expr    act.Clean
	rules   string
	scanner *yr.Scanner
}

// NewCleaner returns cleaner for given env.
func NewCleaner(env *env.Env) *Cleaner {
	return &Cleaner{
		verb:  VerbClean,
		dir:   env.Cfg.Acts.Quarantine.Dir,
		blkSz: int(float64(env.Cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2)),
		expr:  env.Cfg.Acts.Clean,
		rules: env.Paths.Sigs.Yrc,
	}
}

// Load loads given cleaner.
func (cleaner *Cleaner) Load() error {
	if cleaner.dir == "" {
		return acter.ErrDisabled
	}

	rules, err := yr.LoadRules(cleaner.rules)
	if err != nil {
		return fmt.Errorf("%w, %v", sig.ErrYrRulesGet, err)
	}

	cleaner.scanner, err = yr.NewScanner(rules)
	if err != nil {
		return fmt.Errorf("%w, %v", sig.ErrYrScanner, err)
	}

	cleaner.scanner.SetFlags(yr.ScanFlagsFastMode)

	return nil
}

// Act cleans hits for given result.
func (cleaner *Cleaner) Act(result *state.Result) error {
	if cleaner.dir == "" {
		return ErrQuarantineNoDir
	}

	for path, meta := range result.Paths {
		if err := cleaner.clean(path, meta); err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

// clean cleans a given path with given hit meta.
func (cleaner *Cleaner) clean(path string, meta *hit.Meta) error {
	if !filepath.IsAbs(meta.Status) {
		meta.Status = fsys.QuarantinePath(cleaner.dir, path)

		if err := fsys.Mv(path, meta.Status, meta.Attr); err != nil {
			return err
		}
	}

	src, err := os.Open(meta.Status)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, meta.Status)
	}
	defer src.Close()

	dst, err := os.OpenFile(meta.Status+"-clean", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, meta.Status+"-clean")
	}
	defer dst.Close()

	for _, rule := range meta.Rules {
		exprs, ok := cleaner.expr[rule]
		if !ok {
			return fmt.Errorf("%w, %v, %v", ErrCleanNoExpr, rule, meta.Status)
		}

		slog.Info("attempting clean", "rule", rule, "path", meta.Status)

		if err := cleaner.transform(src, dst, rule, exprs); err != nil {
			return err
		}
	}

	slog.Info("clean", "path", dst.Name())

	return fsys.Mv(dst.Name(), path, meta.Attr)
}

// transform executes given sed expr for given src and dst files.
func (cleaner *Cleaner) transform(src *os.File, dst *os.File, rule string, exprs []string) error {
	var (
		matches = &yr.MatchRules{}
		buff    = make([]byte, cleaner.blkSz)
		result  string
	)

	for {
		offset, err := src.Read(buff)

		if offset > 0 {
			for _, expr := range exprs {
				engine, err := sed.New(strings.NewReader(expr))
				if err != nil {
					return err
				}

				result, err = engine.RunString(string(buff[:offset]))
				if err != nil {
					return err
				}

				cleaner.scanner.SetCallback(matches)

				if err := cleaner.scanner.ScanMem([]byte(result)); err != nil {
					return err
				}

				if sig.HasMatch(matches, rule) {
					return fmt.Errorf("%w, %v", ErrCleanFailed, src.Name())
				}
			}

			if _, err := dst.WriteString(result); err != nil {
				return err
			}

			buff = buff[:cleaner.blkSz]
			matches = nil
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}

			break
		}
	}

	return nil
}

// Verb returns a given cleaner verb.
func (cleaner *Cleaner) Verb() string {
	return cleaner.verb
}
