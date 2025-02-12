// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "errors"

var (
	// ErrReqPrep means aequest prepare error.
	ErrReqPrep = errors.New("http: req prepare error")

	// ErrReqDo means request do error.
	ErrReqDo = errors.New("http: req do error")

	// ErrBadStatus means bad status code.
	ErrBadStatus = errors.New("http: bad status code")
)
