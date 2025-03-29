// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package secret

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	var (
		want = t.TempDir()
		got  = New(want)
	)

	if got.path != want {
		t.Errorf("unexpected new cfg result %v, want %v", got.path, want)
	}
}

func TestLoad(t *testing.T) {
	var (
		mock = `[Alerts]
  [Alerts.JSON]
    User = ""
	Pass = ""
	[Alerts.PagerDuty]
	  Token = ""
	[Alerts.SMTP]
	  Hostname = ""
	  Port = 587
	  User = ""
	  Pass = ""

  [Submit]
    Endpoint = ""
    Key = ""

  [S3]
	Endpoint = ""
	Region = ""
	Key = ""
	Secret = ""
  
  [[Git]]
`

		path = filepath.Join(t.TempDir(), t.Name())
	)

	if err := os.WriteFile(path, []byte(mock), 0600); err != nil {
		t.Fatalf("cfg write err: %v", err)
	}

	cfg := &Cfg{
		path: path,
	}

	if err := cfg.Load(); err != nil {
		t.Errorf("cfg load err: %v", err)
	}
}

func TestCfgPath(t *testing.T) {
	var (
		input = &Cfg{
			path: t.Name(),
		}

		want = t.Name()
		got  = input.Path()
	)

	if got != want {
		t.Errorf("unexpected cfg path result %v, want %v", got, want)
	}
}

func TestMock(t *testing.T) {
	if _, err := Mock(t.TempDir()); err != nil {
		t.Errorf("cfg mock err: %v", err)
	}
}
