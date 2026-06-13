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
	verb  string
	dir   string
	blkSz int
	expr  act.Clean
	root  *os.Root
}

type token struct {
	cmd   rune
	flags string
}

// NewCleaner returns cleaner for given env.
func NewCleaner(env *env.Env) *Cleaner {
	return &Cleaner{
		verb: VerbClean,
		dir:  env.Cfg.Acts.Quarantine.Dir,
		// #nosec G404 -- non crypto jitter.
		blkSz: int(float64(env.Cfg.Scans.BlkSz) * (0.8 + rand.Float64()*0.2)),
		expr:  env.Cfg.Acts.Clean,
	}
}

// Load loads given cleaner.
func (cleaner *Cleaner) Load(_ *os.Root) error {
	if cleaner.dir == "" {
		return acter.ErrDisabled
	}

	for rule, exprs := range cleaner.expr {
		for _, expr := range exprs {
			if err := isSafeExpr(expr); err != nil {
				return fmt.Errorf("%w rule %q %v", ErrCleanExprNotSafe, rule, err)
			}
		}
	}

	root, err := os.OpenRoot(cleaner.dir)
	if err != nil {
		return err
	}

	cleaner.root = root

	return nil
}

// Act cleans hits for given result.
func (cleaner *Cleaner) Act(result *state.Result) error {
	if cleaner.dir == "" {
		return ErrQuarantineNoDir
	}

	if cleaner.root == nil {
		return fsys.ErrPathRoot
	}

	sigs, err := sig.Acquire()
	if err != nil {
		return fmt.Errorf("%w, %v", sig.ErrYrcGet, err)
	}
	defer sigs.Release()

	scanner, err := yr.NewScanner(sigs.Rules)
	if err != nil {
		return fmt.Errorf("%w, %v", sig.ErrYrScanner, err)
	}
	defer scanner.Destroy()

	scanner.SetFlags(yr.ScanFlagsFastMode)

	for path, meta := range result.Paths {
		if err := cleaner.clean(path, meta, scanner); err != nil {
			result.AddErr(fmt.Errorf("%w, %v, %v", ErrCleanFail, err, path))
		}
	}

	return nil
}

// clean cleans a given path with given hit meta.
func (cleaner *Cleaner) clean(path string, meta *hit.Meta, scanner *yr.Scanner) error {
	fd, _, err := fsys.Open(path)
	if err != nil {
		return fmt.Errorf("%w, %v", err, path)
	}
	defer fsys.CloseFd(fd)

	meta.Status = fsys.QuarantinePath(cleaner.dir, path)

	if err := fsys.Mv(path, meta.Status, meta.Attr); err != nil {
		return fmt.Errorf("%w, %v", ErrMv, err)
	}

	cleaned, err := cleaner.transform(meta, scanner)
	if err != nil {
		return err
	}

	if err := fsys.Mv(cleaned, path, meta.Attr); err != nil {
		if err := fsys.Unlink(cleaned, false); err != nil {
			slog.Error(ErrDel.Error(), "msg", err, "path", cleaned)
		}

		return fmt.Errorf("%w, %v", ErrMv, err)
	}

	slog.Info("clean", "path", cleaned)

	return nil
}

// transform transforms given meta.
func (cleaner *Cleaner) transform(meta *hit.Meta, scanner *yr.Scanner) (string, error) {
	var (
		clean = filepath.Clean(meta.Status + "-clean")
		ok    bool
	)

	srcName, err := fsys.RootName(cleaner.root, meta.Status)
	if err != nil {
		return "", err
	}

	cleanName, err := fsys.RootName(cleaner.root, clean)
	if err != nil {
		return "", err
	}

	srcF, err := cleaner.root.OpenFile(srcName, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer fsys.Close(srcF)

	dstF, err := cleaner.root.OpenFile(cleanName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer fsys.Close(dstF)

	defer func() {
		if !ok {
			if err := fsys.Unlink(clean, false); err != nil {
				slog.Error(fsys.ErrFileDel.Error(), "msg", err, "path", clean)
			}
		}
	}()

	var engines []*sed.Engine

	for _, rule := range meta.Rules {
		exprs, has := cleaner.expr[rule]
		if !has {
			return "", fmt.Errorf("%w, %v, %v", ErrCleanExprMissing, rule, meta.Status)
		}

		slog.Info("attempting clean", "rule", rule, "path", meta.Status)

		for _, expr := range exprs {
			if err := isSafeExpr(expr); err != nil {
				return "", err
			}

			engine, err := sed.New(strings.NewReader(expr))
			if err != nil {
				return "", fmt.Errorf("%w, %v,", ErrCleanExprCompile, err)
			}

			engines = append(engines, engine)
		}
	}

	if err := cleaner.apply(srcF, dstF, meta.Rules, engines, scanner); err != nil {
		return "", err
	}

	ok = true

	return clean, nil
}

// apply streams src to dst with every engine.
func (cleaner *Cleaner) apply(src *os.File, dst *os.File, rules []string, engines []*sed.Engine, scanner *yr.Scanner) error {
	var (
		matches = &yr.MatchRules{}
		buff    = make([]byte, cleaner.blkSz)
	)

	for {
		offset, err := src.Read(buff)

		if offset > 0 {
			data := string(buff[:offset])

			for _, engine := range engines {
				data, err = engine.RunString(data)
				if err != nil {
					return fmt.Errorf("%w, %v,", ErrCleanExprDo, err)
				}
			}

			*matches = yr.MatchRules{}

			scanner.SetCallback(matches)

			if err = scanner.ScanMem([]byte(data)); err != nil {
				return err
			}

			for _, rule := range rules {
				if sig.HasMatch(matches, rule) {
					return fmt.Errorf("%w, %v", ErrCleanFail, src.Name())
				}
			}

			if _, err = dst.WriteString(data); err != nil {
				return fmt.Errorf("%w, %v", ErrCleanFail, dst.Name())
			}
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

func (cleaner *Cleaner) release() {
	if cleaner.root != nil {
		cleaner.root.Close()
		cleaner.root = nil
	}
}

// Verb returns given cleaner verb.
func (cleaner *Cleaner) Verb() string {
	return cleaner.verb
}
