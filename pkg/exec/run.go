// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// Run executes given binary with given args.
// Mainly for integrations, nothing else. Avoid this wherever possible.
func Run(bin string, args ...string) (string, error) {
	if slices.ContainsFunc(append(args, bin), func(input string) bool {
		return reMetaChars.MatchString(input)
	}) {
		return "", ErrMetaChars
	}

	var (
		cmd    = exec.Command(bin, args...)
		stdout = &strings.Builder{}
		stderr = &strings.Builder{}
	)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%w, %v, %v, %v", ErrRun, err, "stderr", stderr.String())
	}

	return stdout.String(), nil
}
