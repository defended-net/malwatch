// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package info

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

// Report represents the ver plus cfg and platform info.
type Report struct {
	Ver  string `json:"ver"`
	Go   string `json:"go"`
	Cfgs *Cfgs  `json:"core"`
	Plat string `json:"platform"`
}

// Cfgs represents cfg paths
type Cfgs struct {
	Config string `json:"cfg"`
	Acts   string `json:"actions"`
}

// Do displays ver info.
// ./malwatch info
func Do(env *env.Env, _ []string) error {
	info, _ := debug.ReadBuildInfo()

	report := &Report{
		Ver: env.Ver,
		Go:  info.GoVersion,

		Cfgs: &Cfgs{
			Config: env.Paths.Cfg.Base,
			Acts:   env.Paths.Cfg.Acts,
		},

		Plat: env.Plat.Cfg().Path(),
	}

	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}
