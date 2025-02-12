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
)

// No actual coverage here, we just want to make sure everything is accurately mapped.
func TestLayout(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  string
	}{
		"start": {
			input: []string{
				"start",
			},

			want: "monitor.Do",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = append([]string{""}, test.input...)
			flag.Parse()

			unwrapped, err := cmds.Unwrap()
			if err != nil && !strings.HasSuffix(err.Error(), test.want) {
				t.Fatalf("unwrap error: %v", err)
			}

			if unwrapped == nil {
				return
			}

			result := runtime.FuncForPC(reflect.ValueOf(unwrapped.Fn).Pointer()).Name()

			if !strings.HasSuffix(result, test.want) {
				t.Errorf("unexpected cli result %v, want %v", result, test.want)
			}
		})
	}
}
