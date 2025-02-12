// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package secret

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/defended-net/malwatch/pkg/fsys"
)

// Cfg represents secrets cfg.
// path is toml path.
// tests require respective env vars to run.
type Cfg struct {
	path   string
	Alerts *Alerts
	S3     *S3
	Git    []*Repo
}

// Alerts represents alerts.
type Alerts struct {
	JSON      *JSON
	PagerDuty *PagerDuty
	SMTP      *SMTP
}

// JSON represents json.
type JSON struct {
	User string `env:"JSON_USER"`
	Pass string `env:"JSON_PASS"`
}

// PagerDuty represents pagerduty.
type PagerDuty struct {
	Token string `env:"PD_TOKEN"`
}

// SMTP represents smtp.
type SMTP struct {
	Hostname string `env:"SMTP_HOSTNAME"`
	Port     int    `env:"SMTP_PORT"`
	User     string `env:"SMTP_USER"`
	Pass     string `env:"SMTP_PASS"`
}

// S3 represents s3.
type S3 struct {
	Endpoint string `env:"S3_ENDPOINT"`
	Region   string `env:"S3_REGION"`
	Bucket   string `env:"S3_BUCKET"`
	Key      string `env:"S3_KEY"`
	Secret   string `env:"S3_SECRET"`
}

// Repo represents a git repo.
type Repo struct {
	User  string
	Token string
	URL   string
}

// New returns a new cfg.
func New(path string) *Cfg {
	return &Cfg{
		path: path,

		Alerts: &Alerts{
			JSON:      &JSON{},
			PagerDuty: &PagerDuty{},
			SMTP:      &SMTP{},
		},

		S3: &S3{},
	}
}

// Load reads the cfg from toml path.
func (cfg *Cfg) Load() error {
	if err := fsys.ReadTOML(cfg.Path(), cfg); err != nil {
		return err
	}

	for idx, repo := range cfg.Git {
		if strings.TrimSuffix(repo.URL, "/") != "https://github.com/defended-net/malwatch-signatures" &&
			repo.Token == "" {
			repo.User = os.Getenv("GIT_" + fmt.Sprint(idx) + "_USER")
			repo.Token = os.Getenv("GIT_" + fmt.Sprint(idx) + "_TOKEN")
			repo.URL = os.Getenv("GIT_" + fmt.Sprint(idx) + "_URL")
		}
	}

	return nil
}

// Path returns a given cfg's toml path.
func (cfg *Cfg) Path() string {
	return cfg.path
}

// Mock mocks a cfg.
func Mock(path string) (*Cfg, error) {
	cfg := &Cfg{
		path: path,

		Alerts: &Alerts{
			JSON:      &JSON{},
			PagerDuty: &PagerDuty{},
			SMTP:      &SMTP{},
		},

		S3: &S3{},

		Git: []*Repo{
			{
				URL: "https://github.com/defended-net/malwatch-signatures",
			},
		},
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	for idx, repo := range cfg.Git {
		if repo.Token == "" {
			repo.User = os.Getenv("GIT_" + fmt.Sprint(idx) + "_USER")
			repo.Token = os.Getenv("GIT_" + fmt.Sprint(idx) + "_TOKEN")
			repo.URL = os.Getenv("GIT_" + fmt.Sprint(idx) + "_URL")
		}
	}

	return cfg, nil
}
