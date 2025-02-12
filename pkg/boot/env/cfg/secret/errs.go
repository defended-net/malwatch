// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package secret

import "errors"

var (
	// ErrAWSNoRegion means no region in cfg.
	ErrAWSNoRegion = errors.New("secret: region not configured")

	// ErrAWSNoCreds means no creds in cfg.
	ErrAWSNoCreds = errors.New("secret: credentials not configured")
)
