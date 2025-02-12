// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

// Cfg represents scan cfg.
// BlkSz is a fixed size for chunked file reads.
// BatchSz is the maximum number of detections per alert before the next is created.
type Cfg struct {
	Targets []string
	Paths   []string
	Timeout int `env:"TIMEOUT"`
	MaxAge  int `env:"MTIME"`
	BlkSz   int `env:"BLOCK_SZ"`
	BatchSz int `env:"BATCH_SZ"`
	Monitor *Monitor
}

// Monitor represents monitor scan cfg.
// Timeout allows for ongoing file changes to continue before scan.
type Monitor struct {
	Timeout int `env:"MONITOR_TIMEOUT"`
}

// Quarantine represents quarantine cfg.
type Quarantine struct {
	Dir string `env:"QUAR_DIR"`
}

// New returns a new cfg. Populated for install defaults.
func New() *Cfg {
	return &Cfg{
		Targets: []string{
			`^/var/www/(?P<target>[^/]+)`,
		},

		Paths: []string{
			"/var/www",
		},

		Timeout: 60,
		MaxAge:  0,

		BlkSz:   65536,
		BatchSz: 500,

		Monitor: &Monitor{
			Timeout: 5,
		},
	}
}
