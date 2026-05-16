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
  [Alerts.Slack]
    Webhook = ""
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

		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())

		input = &Cfg{
			path: path,
		}
	)

	if err := os.WriteFile(path, []byte(mock), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); got != nil {
		t.Errorf("load err %v", got)
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
		t.Errorf("unexpected path result %v, want %v", got, want)
	}
}

func TestMock(t *testing.T) {
	if _, got := Mock(t.TempDir()); got != nil {
		t.Errorf("cfg mock err %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	var (
		tmp = t.TempDir()

		cfg = &Cfg{
			path: filepath.Join(tmp, t.Name()),
		}
	)

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := cfg.Load(root); got == nil {
		t.Errorf("unexpected load success")
	}
}

func TestLoadGitURLEnv(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())

		got = &Cfg{
			path: path,
		}

		mock = `[Alerts]
  [Alerts.JSON]
    User = ""
    Pass = ""

  [Submit]
    Endpoint = ""
    Key = ""

  [S3]
    Endpoint = ""
    Region = ""

[[Git]]
URL = "https://github.com/example/private"
`

		err = os.WriteFile(path, []byte(mock), 0600)
	)

	if err != nil {
		t.Fatalf("cfg write err %v", err)
	}

	t.Setenv("GIT_0_USER", "u")
	t.Setenv("GIT_0_TOKEN", "tok")
	t.Setenv("GIT_0_URL", "https://github.com/example/private")

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if err := got.Load(root); err != nil {
		t.Errorf("load err %v", err)
	}

	if got.Git[0].User != "u" {
		t.Errorf("unexpected git user %v", got.Git[0].User)
	}
}

func TestMockGitURLEnv(t *testing.T) {
	t.Setenv("GIT_0_USER", "u")
	t.Setenv("GIT_0_TOKEN", "tok")
	t.Setenv("GIT_0_URL", "https://github.com/example/private")

	got, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("mock err %v", err)
	}

	if got.Git[0].User != "" {
		t.Errorf("unexpected git user override %v", got.Git[0].User)
	}
}

func TestMockEnvErr(t *testing.T) {
	t.Setenv("SMTP_PORT", "invalid")

	if _, got := Mock(filepath.Join(t.TempDir(), t.Name())); got == nil {
		t.Errorf("unexpected mock success")
	}
}
