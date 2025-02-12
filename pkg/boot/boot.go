// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package boot

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg"
	"github.com/defended-net/malwatch/pkg/boot/env/plat"
	"github.com/defended-net/malwatch/pkg/boot/install"
	"github.com/defended-net/malwatch/pkg/cmd"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/logger"
)

var tasks = []func(*env.Env) error{
	install.Run,
	cfg.Load,
	logger.Load,
	db.Load,
	plat.Load,
}

// Run starts boot based on tasks.
func Run(env *env.Env) error {
	// Basic logging to start.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := cmd.Listen(env.State); err != nil {
		return err
	}

	for _, task := range tasks {
		if err := task(env); err != nil {
			return err
		}
	}

	runtime.GOMAXPROCS(env.Cfg.Cores)

	return nil
}
