// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

// Cfg represents acts cfg. path stores the toml filepath.
// Default stores the default acts initially applicable to all detections.
type Cfg struct {
	path       string
	Default    []string
	Signatures map[string][]string
	Paths      map[string]map[string][]string
	Skips      *Skips
	Quarantine *Quarantine
	Clean      Clean
}

// Skips represents dir and file path skips.
// Files serves as a dict.
type Skips struct {
	Dirs  []string
	Files map[string]struct{}
}

// Quarantine represents the quarantine cfg.
type Quarantine struct {
	Dir string
}

// Clean are sed compatible cleaning expressions.
// name:expressions
type Clean map[string][]string

// Loadout represents rule:acts.
type Loadout struct {
	Rule    string   `json:"rule"`
	Actions []string `json:"actions"`
}

// Validate dedupes and validates verbs.
func Validate(acters []acter.Acter, verbs []string) []string {
	if slices.Equal(verbs, []string{""}) {
		return []string{}
	}

	valid := []string{}

	for _, verb := range verbs {
		if err, _ := acter.Get(acters, verb); err != nil {
			valid = append(valid, verb)
		}
	}

	return valid
}

// New returns a new cfg from given path.
func New(path string) *Cfg {
	return &Cfg{
		path: path,

		Signatures: map[string][]string{},
		Paths:      map[string]map[string][]string{},

		Quarantine: &Quarantine{},
		Clean:      Clean{},
	}
}

// Load reads the cfg from toml path.
func (cfg *Cfg) Load() error {
	return fsys.ReadTOML(cfg.Path(), cfg)
}

// NewVerbs returns verbs based on given path and rules.
// Switch is for probability, for instance a filepath:rule is
// more likely than dir:rule.
func (cfg *Cfg) NewVerbs(path string, rules ...string) []string {
	verbs := []string{}

	path, _ = strings.CutSuffix(path, "/")

	for _, rule := range rules {
		switch {
		// file:rule cfg. Most likely.
		case len(cfg.Paths[path][rule]) > 0:
			verbs = append(verbs, cfg.Paths[path][rule]...)

		// file:rule whitelist. Very likely.
		case cfg.Paths[path][rule] != nil && len(cfg.Paths[path][rule]) == 0:
			verbs = append(verbs, []string{}...)

		// cwd:rule cfg. Very likely.
		case len(cfg.Paths[filepath.Dir(path)][rule]) > 0:
			verbs = append(verbs, cfg.Paths[filepath.Dir(path)][rule]...)

		// cwd:rule whitelist. Likely.
		case cfg.Paths[filepath.Dir(path)][rule] != nil && len(cfg.Paths[filepath.Dir(path)][rule]) == 0:
			verbs = append(verbs, []string{}...)

		// file:rule cfg. Less likely.
		case len(cfg.Paths[path]["*"]) > 0:
			verbs = append(verbs, cfg.Paths[path]["*"]...)

		// cwd:rule cfg. Less likely.
		case len(cfg.Paths[filepath.Dir(path)]["*"]) > 0:
			verbs = append(verbs, cfg.Paths[filepath.Dir(path)]["*"]...)

		// Rule whitelist.
		case cfg.Signatures[rule] != nil && len(cfg.Signatures[rule]) == 0:
			verbs = append(verbs, []string{}...)

		// Rule cfg.
		case len(cfg.Signatures[rule]) > 0:
			verbs = append(verbs, cfg.Signatures[rule]...)

		default:
			verbs = append(verbs, cfg.Default...)
		}
	}

	slices.Sort(verbs)

	return slices.Compact(verbs)
}

// Get returns loadouts for given path.
func (cfg *Cfg) Get(key string) []*Loadout {
	loadouts := []*Loadout{}

	path, _ := strings.CutSuffix(key, "/")

	if rules, ok := cfg.Paths[path]; ok {
		for rule, acts := range rules {
			loadout := &Loadout{
				Rule:    rule,
				Actions: acts,
			}

			loadouts = append(loadouts, loadout)
		}
	}

	return loadouts
}

// AddSigVerbs adds a given sig's verbs.
func (cfg *Cfg) AddSigVerbs(acters []acter.Acter, rule string, verbs []string) error {
	valid := Validate(acters, verbs)

	switch {
	case !re.IsValidYrName(rule):
		return fmt.Errorf("%w, %v", ErrInvalidChars, rule)

	case rule == "*":
		return ErrStarSkipNotAllowed

	case len(valid) == 0 && !slices.Equal(verbs, []string{""}):
		return fmt.Errorf("%w, %v", ErrUnknownVerb, verbs)

	case cfg.Signatures[rule] == nil:
		cfg.Signatures[rule] = []string{}
	}

	slog.Info("setting actions", "input", valid, "rule", rule)

	for _, verb := range valid {
		if !slices.Contains(cfg.Signatures[rule], verb) {
			cfg.Signatures[rule] = append(cfg.Signatures[rule], verb)
		}
	}

	// Otherwise it is a skip.
	if slices.Equal(verbs, []string{""}) {
		cfg.Signatures[rule] = []string{}
	}

	if err := cfg.Compact("", rule); err != nil {
		return err
	}

	return fsys.WriteTOML(cfg.path, cfg)
}

// AddPathVerbs adds given path's verbs.
func (cfg *Cfg) AddPathVerbs(acters []acter.Acter, path string, rule string, verbs []string) error {
	valid := Validate(acters, verbs)

	path, _ = strings.CutSuffix(path, "/")

	switch {
	case !filepath.IsAbs(path):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, path)

	case len(valid) == 0 && !slices.Equal(verbs, []string{""}):
		return fmt.Errorf("%w, %v", ErrUnknownVerb, verbs)

	case !re.IsValidYrName(rule):
		return fmt.Errorf("%w, %v", ErrInvalidChars, rule)

	case cfg.Paths[path] == nil ||
		rule == "*" ||
		cfg.Paths[path]["*"] != nil:
		cfg.Paths[path] = map[string][]string{}
	}

	slog.Info("setting actions", "input", valid, "path", path, "rule", rule)

	for _, verb := range valid {
		if !slices.Contains(cfg.Paths[path][rule], verb) {
			cfg.Paths[path][rule] = append(cfg.Paths[path][rule], verb)
		}
	}

	// Otherwise it is a skip.
	if slices.Equal(verbs, []string{""}) {
		cfg.Paths[path][rule] = []string{}
	}

	if err := cfg.Compact(path, rule); err != nil {
		return err
	}

	return fsys.WriteTOML(cfg.path, cfg)
}

// SetSigVerbs sets given sig's verbs.
func (cfg *Cfg) SetSigVerbs(acters []acter.Acter, rule string, verbs []string) error {
	switch {
	case !re.IsValidYrName(rule):
		return fmt.Errorf("%w, %v", ErrInvalidChars, rule)

	case cfg.Signatures[rule] == nil:
		cfg.Signatures[rule] = []string{}
	}

	cfg.Signatures[rule] = []string{}

	return cfg.AddSigVerbs(acters, rule, verbs)
}

// SetPathVerbs sets given path's verbs.
func (cfg *Cfg) SetPathVerbs(acters []acter.Acter, path string, rule string, verbs []string) error {
	path, _ = strings.CutSuffix(path, "/")

	switch {
	case !filepath.IsAbs(path):
		return fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, path)

	case !re.IsValidYrName(rule):
		return fmt.Errorf("%w, %v", ErrInvalidChars, rule)

	case cfg.Paths[path] == nil:
		cfg.Paths[path] = map[string][]string{}
	}

	cfg.Paths[path][rule] = []string{}

	return cfg.AddPathVerbs(acters, path, rule, verbs)
}

// DelSigVerbs deletes sig's verbs.
func (cfg *Cfg) DelSigVerbs(sig string) error {
	if _, ok := cfg.Signatures[sig]; !ok {
		return fmt.Errorf("%w, %v", ErrNoActs, sig)
	}

	slog.Info("deleting actions", "rule", sig)

	delete(cfg.Signatures, sig)

	return fsys.WriteTOML(cfg.path, cfg)
}

// DelPathVerbs deletes given path's verbs.
func (cfg *Cfg) DelPathVerbs(path string) error {
	path, _ = strings.CutSuffix(path, "/")

	if _, ok := cfg.Paths[path]; !ok {
		return fmt.Errorf("%w, %v", ErrNoActs, path)
	}

	slog.Info("deleting actions", "path", path)

	delete(cfg.Paths, path)

	return fsys.WriteTOML(cfg.path, cfg)
}

// Compact compacts verbs.
// TODO
func (cfg *Cfg) Compact(path string, rule string) error {
	path, _ = strings.CutSuffix(path, "/")

	if cfg.Paths[path] != nil {
		if _, ok := cfg.Paths[path][rule]; ok {
			slices.Sort(cfg.Paths[path][rule])
			cfg.Paths[path][rule] = slices.Compact(cfg.Paths[path][rule])
		}
	}

	if cfg.Signatures != nil {
		if _, ok := cfg.Signatures[rule]; ok {
			slices.Sort(cfg.Signatures[rule])
			cfg.Signatures[rule] = slices.Compact(cfg.Signatures[rule])
		}
	}

	return nil
}

// Path returns the cfg file path.
func (cfg *Cfg) Path() string {
	return cfg.path
}

// GetSkips returns dir and file path skips from a given cfg.
func GetSkips(cfg *Cfg, paths *path.Paths) *Skips {
	dirs := []string{
		paths.Install.Dir,
		cfg.Quarantine.Dir,
		filepath.Dir(paths.Install.Db),
		filepath.Dir(paths.Install.Log),
	}

	skips := &Skips{
		Files: map[string]struct{}{},
	}

	for _, path := range dirs {
		if path != "" && path != "." {
			skips.Dirs = append(skips.Dirs, path)
		}
	}

	for path, sigs := range cfg.Paths {
		if _, ok := sigs["*"]; !ok || !slices.Equal(sigs["*"], []string{}) {
			continue
		}

		stat, err := os.Stat(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				slog.Error(fsys.ErrStat.Error(), "path", path)
			}

			continue
		}

		if stat.IsDir() {
			skips.Dirs = append(skips.Dirs, strings.TrimSuffix(path, "/"))
			continue
		}

		skips.Files[path] = struct{}{}
	}

	return skips
}

// Mock mocks a cfg.
func Mock(path string) *Cfg {
	return &Cfg{
		path:       path,
		Signatures: map[string][]string{},
		Paths:      map[string]map[string][]string{},

		Quarantine: &Quarantine{
			Dir: filepath.Dir(path),
		},

		Clean: Clean{},
	}
}
