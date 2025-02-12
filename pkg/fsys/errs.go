// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import "errors"

// FILES
var (
	// ErrFileOpen means file open error.
	ErrFileOpen = errors.New("fsys: file open error")

	// ErrFileCreate means file create error.
	ErrFileCreate = errors.New("fsys: file create error")

	// ErrFileCopy means file copy error.
	ErrFileCopy = errors.New("fsys: file copy error")

	// ErrFileDel means file del error.
	ErrFileDel = errors.New("fsys: file del error")
)

// DIR
var (
	// ErrWalk means walk error.
	ErrWalk = errors.New("fsys: walk error")

	// ErrDirCreate means dir create error.
	ErrDirCreate = errors.New("fsys: dir create error")

	// ErrDirMv means dir move attempt.
	ErrDirMv = errors.New("fsys: only files permitted for move")
)

// MODE
var (
	// ErrStat means stat error.
	ErrStat = errors.New("fsys: stat error")

	// ErrChmod means chmod error.
	ErrChmod = errors.New("fsys: chmod error")

	// ErrChown means chown error.
	ErrChown = errors.New("fsys: chown error")
)

// VALIDATION
var (
	// ErrPathInvalid means invalid path format.
	ErrPathInvalid = errors.New("fsys: invalid path format")

	// ErrPathRoot means root path not permitted.
	ErrPathRoot = errors.New("fsys: root path not permitted")

	// ErrPathNotAbs means path not absolute.
	ErrPathNotAbs = errors.New("fsys: path not absolute")
)

// TOML
var (
	// ErrTOMLRead means toml read error.
	ErrTOMLRead = errors.New("fsys: toml read error")

	// ErrTOMLWrite means toml write error.
	ErrTOMLWrite = errors.New("fsys: toml write error")
)
