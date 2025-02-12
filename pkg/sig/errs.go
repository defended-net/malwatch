// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import "errors"

// YR

var (
	// ErrYrScanner means yr scanner error.
	ErrYrScanner = errors.New("sig: yr scanner error")

	// ErrYrCompiler means yr compiler error.
	ErrYrCompiler = errors.New("sig: yr compiler error")

	// ErrYrIdxAdd means yr index add error.
	ErrYrIdxAdd = errors.New("sig: yr index add error")

	// ErrYrIdxWrite means yr index write error.
	ErrYrIdxWrite = errors.New("sig: yr index write error")

	// ErrYrRulesLoad means yr load rules error.
	ErrYrRulesLoad = errors.New("sig: yr load rules error")

	// ErrYrRulesGet means yr get rules error.
	ErrYrRulesGet = errors.New("sig: yr get rules error")

	// ErrYrRulesSave means yr save rules error.
	ErrYrRulesSave = errors.New("sig: yr save rules error")
)

// GIT

var (
	// ErrNoRepos means no repos in cfg.
	ErrNoRepos = errors.New("sig: no repos in config")

	// ErrNoRepoName means repo name is 0 length.
	ErrNoRepoName = errors.New("sig: 0 length repo name")

	// ErrNoRepoOwner means the repo owner is 0 length.
	ErrNoRepoOwner = errors.New("sig: 0 length repo owner")
)

// FS

var (
	// ErrNoSigTmpDir means sig tmp dir is 0 length.
	ErrNoSigTmpDir = errors.New("sig: 0 length sig tmp dir")

	// ErrIsSym means symlink found.
	ErrIsSym = errors.New("sig: symlinks not permitted")
)
