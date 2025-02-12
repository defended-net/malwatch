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
		t.Errorf("env mock error: %v", err)
	}

	if err := sig.Mock(_env); err != nil {
		t.Errorf("sig mock error: %v", err)
	}

	go func(env *env.Env) {
		if err := Do(env, []string{}); !errors.Is(err, context.Canceled) {
			t.Errorf("start error: %v", err)
		}
	}(_env)

	// Allow time for monitor to load.
	time.Sleep(3 * time.Second)

	_env.State.GetCancels()[0]()
}
