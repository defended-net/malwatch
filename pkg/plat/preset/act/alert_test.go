// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestAlertVerb(t *testing.T) {
	input := &Alerter{
		verb: VerbAlert,
	}

	if got := input.Verb(); got != VerbAlert {
		t.Errorf("unexpected verb result %v, want %v", got, VerbAlert)
	}
}

func TestActAlert(t *testing.T) {
	tests := map[string]struct {
		input *state.Result
		want  error
	}{
		"single": {
			input: state.NewResult("",
				state.Paths{
					t.Name(): {},
				}),

			want: nil,
		},

		"multi": {
			input: state.NewResult("",
				state.Paths{
					t.Name():        {},
					t.Name() + "-b": {},
				}),

			want: nil,
		},

		"none": {
			input: state.NewResult("",
				state.Paths{}),

			want: nil,
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	alerter := NewAlerter(env)

	for _, alerter := range alerter.senders {
		path := alerter.Cfg().Path()

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatalf("dir create error: %s", err)
		}

		if _, err := os.Create(path); err != nil {
			t.Fatalf("file create error: %s", err)
		}
	}

	if err := alerter.Load(); err != nil {
		t.Fatalf("alerter load error: %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := alerter.Act(test.input); err != nil && !strings.HasSuffix(err.Error(), "connection refused") {
				t.Errorf("alerter error: %v", err)
			}
		})
	}
}

func TestAlerterLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	alerter := NewAlerter(env)

	for _, alerter := range alerter.senders {
		path := alerter.Cfg().Path()

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatalf("dir create error: %s", err)
		}

		if _, err := os.Create(path); err != nil {
			t.Fatalf("file create error: %s", err)
		}
	}

	if err := alerter.Load(); err != nil {
		t.Errorf("alerter load error: %v", err)
	}
}

func TestAlerterLoadErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %s", err)
	}

	env.Paths.Alerts.Dir = "/dev/null"

	alerter := NewAlerter(env)

	if err := alerter.Load(); !errors.Is(err, ErrCfgLoad) {
		t.Errorf("unexpected alerter load error: %s", err)
	}
}
