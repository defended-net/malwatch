// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

// Cfg represents log cfg.
// Verbose adds src line as field.
type Cfg struct {
	Dir     string `env:"LOG_DIR"`
	Verbose bool   `env:"LOG_VERBOSE"`
}

// Mock mocks a cfg.
func Mock(path string) *Cfg {
	return &Cfg{
		Dir:     path,
		Verbose: true,
	}
}
