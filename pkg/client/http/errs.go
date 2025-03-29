// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "errors"

var (
	// ErrReqPrep means request prepare err.
	ErrReqPrep = errors.New("http: req prepare error")

	// ErrReqDo means request do err.
	ErrReqDo = errors.New("http: req do error")

	// ErrBadStatus means bad status code.
	ErrBadStatus = errors.New("http: bad status code")

	// ErrSubmit means malware sample submit err.
	ErrSubmit = errors.New("http: submit malware sample error")
)
