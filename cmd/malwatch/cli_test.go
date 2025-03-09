// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/cli/help"
)

// No actual coverage here, we just want to make sure everything is accurately mapped.
func TestLayout(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  string
	}{
		"scan": {
			input: []string{
				"scan",
			},

			want: "scan.Do",
		},

		"history": {
			input: []string{
				"history",
			},

			want: help.ErrHistory.Error(),
		},

		"actions": {
			input: []string{
				"actions",
			},
			want: help.ErrActs.Error(),
		},

		"actions get": {
			input: []string{
				"actions",
				"get",
			},

			want: help.ErrActs.Error(),
		},

		"actions set": {
			input: []string{
				"actions",
				"set",
			},

			want: help.ErrActs.Error(),
		},

		"actions del": {
			input: []string{
				"actions",
				"del",
			},

			want: help.ErrActs.Error(),
		},

		"quarantine": {
			input: []string{
				"quarantine",
			},

			want: help.ErrQuarantine.Error(),
		},

		"restore": {
			input: []string{
				"restore",
			},

			want: help.ErrRestore.Error(),
		},

		"signatures": {
			input: []string{
				"signatures",
			},

			want: help.ErrSigs.Error(),
		},

		"signatures update": {
			input: []string{
				"signatures",
				"update",
			},

			want: string("sig.Update"),
		},

		"signatures refresh": {
			input: []string{
				"signatures",
				"refresh",
			},

			want: string("sig.Refresh"),
		},

		"info": {
			input: []string{
				"info",
			},

			want: "info.Do",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = append([]string{""}, test.input...)
			flag.Parse()

			cmd, err := cmds.Route()
			if err != nil && !strings.HasSuffix(err.Error(), test.want) {
				t.Fatalf("route error: %v", err)
			}

			if cmd == nil {
				return
			}

			result := runtime.FuncForPC(reflect.ValueOf(cmd.Fn).Pointer()).Name()

			if !strings.HasSuffix(result, test.want) {
				t.Errorf("unexpected cli result %v, want %v", result, test.want)
			}
		})
	}
}
