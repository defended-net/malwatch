// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/sig"
)

var sample = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`

func TestMain(m *testing.M) {
	if os.Getuid() != 0 {
		fmt.Println("monitor: tests require root")
		return
	}

	m.Run()
}

func TestNewErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Paths.Sigs.Yrc = ""

	if _, got := New(env); !errors.Is(got, sig.ErrYrRulesLoad) {
		t.Errorf("unexpected new monitor error: %v, want %v", got, sig.ErrYrRulesLoad)
	}
}

func TestRun(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, t.Name())
	)

	re.SetTargets(regexp.MustCompile(`^/(?P<target>[^/]+)`))

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create error: %v", err)
	}

	go func() {
		time.Sleep(3 * time.Second)

		for _, file := range []struct {
			path string
			data []byte
		}{
			{
				path, []byte(sample),
			},
			{
				path, []byte(sample),
			},
			{
				filepath.Join(dir, "clean"), []byte(`clean`),
			},
		} {
			if err := os.WriteFile(file.path, file.data, 0600); err != nil {
				t.Errorf("file write error: %v", err)
			}
		}

		time.Sleep(10 * time.Second)

		env.State.GetCancels()[0]()
	}()

	if err := Run(env); !errors.Is(err, context.Canceled) {
		t.Errorf("run error: %v", err)
	}
}
