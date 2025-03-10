// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/monitor"
)

// Do starts a file monitor.
// ./malwatch monitor start
func Do(env *env.Env, _ []string) error {
	return monitor.Run(env)
}
