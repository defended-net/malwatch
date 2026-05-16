// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/sig"
)

func TestStart(t *testing.T) {
	if os.Getuid() != 0 {
		fmt.Println("monitor: tests require root")

		return
	}

	_env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if err := sig.Mock(_env, true); err != nil {
		t.Errorf("sig mock err %v", err)
	}

	go func(env *env.Env) {
		if got := Do(env, []string{}); !errors.Is(got, context.Canceled) {
			t.Errorf("do err %v", got)
		}
	}(_env)

	// Allow time for monitor to load.
	time.Sleep(3 * time.Second)

	_env.State.GetCancels()[0]()
}
