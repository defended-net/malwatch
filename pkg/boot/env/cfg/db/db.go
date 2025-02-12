// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

// Cfg represents db cfg.
type Cfg struct {
	Dir string `env:"DB_DIR"`
}
