// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/yr"
)

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		t.Fatalf("yara load error: %v", err)
	}

	if _, err := New(env.Cfg, rules); err != nil {
		t.Errorf("worker create error: %v", err)
	}
}

func TestWork(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	queue := make(chan string)

	go func() {
		defer close(queue)
		queue <- filepath.Join(t.TempDir(), t.Name())
	}()

	state := state.NewJob()

	state.WGrp.Add(1)

	worker.Work(context.Background(), state, queue)
}

func TestScan(t *testing.T) {
	var (
		rule = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`
		path = filepath.Join(t.TempDir(), t.Name())
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	if err := os.WriteFile(path, []byte(rule), 0600); err != nil {
		t.Errorf("file write error: %v", err)
	}

	state := state.NewJob()

	go func(worker *Worker) {
		defer close(state.Hits)

		worker.Scan(path, state)
	}(worker)
}

func TestMatchesToString(t *testing.T) {
	tests := map[string]struct {
		input yr.MatchRules
		want  []string
	}{
		"single": {
			input: yr.MatchRules{
				{
					Rule: "rule",
				},
			},

			want: []string{"rule"},
		},

		"compound": {
			input: yr.MatchRules{
				{
					Rule: "rule-a",
				},
				{
					Rule: "rule-b",
				},
			},

			want: []string{
				"rule-a",
				"rule-b",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := MatchesToStr(test.input)

			if !slices.Equal(result, test.want) {
				t.Errorf("unexpected matches to string result %v, want %v", result, test.want)
			}
		})
	}

}

func TestMock(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if _, err := Mock(env); err != nil {
		t.Errorf("sig mock error: %v", err)
	}
}
