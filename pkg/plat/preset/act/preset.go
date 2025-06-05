package act

import (
	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

// Preset returns preset acts for given env.
func Preset(env *env.Env) []acter.Acter {
	return []acter.Acter{
		NewExiler(env),
		NewQuarantiner(env),
		NewCleaner(env),
		NewAlerter(env),
	}
}
