// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
)

// Parsed represents parsed args.
type Parsed struct {
	key   string
	path  string
	rule  string
	verbs []string
}

// Get prints acts for given target or path.
// ./malwatch actions get [PATH | SIGNATURE]
func Get(env *env.Env, args []string) error {
	var (
		parsed   = Parse(args)
		loadouts = env.Cfg.Acts.Get(parsed.key)
	)

	json, err := json.MarshalIndent(loadouts, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(json))

	return nil
}

// Set sets acts for path or rule name.
// ./malwatch acts set [PATH | RULE]
func Set(env *env.Env, args []string) error {
	var (
		parsed = Parse(args)
		acters = env.Plat.Acters()
	)

	if parsed.path != "" {
		return env.Cfg.Acts.SetPathVerbs(acters, parsed.path, parsed.rule, parsed.verbs)
	}

	return env.Cfg.Acts.SetSigVerbs(acters, parsed.key, parsed.verbs)
}

// Del deletes acts for path or rule name.
// ./malwatch actions del [PATH | SIGNATURE]
func Del(env *env.Env, args []string) error {
	parsed := Parse(args)

	if parsed.path != "" {
		return env.Cfg.Acts.DelPathVerbs(parsed.path)
	}

	return env.Cfg.Acts.DelSigVerbs(parsed.key)
}

// Parse parses cli args. A key from first arg determines if path or rule name was specified.
func Parse(args []string) *Parsed {
	parsed := &Parsed{
		key:   args[0],
		rule:  re.FindYrName(args[0]),
		verbs: []string(args[1:]),
	}

	if filepath.IsAbs(parsed.key) {
		parsed.path = parsed.key
	}

	if len(args) > 1 && parsed.rule == "" {
		parsed.rule = re.FindYrName(args[1])
		parsed.verbs = parsed.verbs[1:]
	}

	// path and rule whitelist or skip
	if len(args) == 1 || len(parsed.verbs) == 0 {
		parsed.verbs = []string{
			"",
		}
	}

	// path skip
	if len(args) == 1 && parsed.path != "" {
		parsed.rule = "*"
	}

	return parsed
}
