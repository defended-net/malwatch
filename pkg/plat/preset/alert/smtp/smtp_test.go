// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"fmt"
	"os"
	"testing"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestMain(m *testing.M) {
	if os.Getenv("SMTP_HOSTNAME") == "" {
		fmt.Println("smtp: tests require env var")
		return
	}

	m.Run()
}

func TestAlert(t *testing.T) {
	sender, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("sender mock error: %v", err)
	}

	if got := sender.Alert(&state.Result{}); got != nil {
		t.Errorf("unexpected alert error: %v", got)
	}
}
