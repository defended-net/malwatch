// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

// Plat represents a cpanel platform.
type Plat struct {
	env        *env.Env
	cfg        *Cfg
	acters     []acter.Acter
	bin        string
	domainInfo []string
}

// DomainInfo represents an get_domain_info api response.
// https://api.docs.cpanel.net/openapi/whm/operation/get_domain_info
type DomainInfo struct {
	Data struct {
		Domains []Domain `json:"domains"`
	} `json:"data"`

	Meta Meta `json:"metadata"`
}

// Domain represents the domain portion of get_domain_info responses.
type Domain struct {
	User    string `json:"user"`
	Domain  string `json:"domain"`
	Docroot string `json:"docroot"`
}

// AccSummary represents an account summary response.
// https://api.docs.cpanel.net/openapi/whm/operation/accountsummary
type AccSummary struct {
	Data struct {
		Acc []Acc `json:"acct"`
	} `json:"data"`

	Meta Meta `json:"metadata"`
}

// Acc represents the account portion of accountsummary responses.
type Acc struct {
	Email string `json:"email"`
}

// Meta represents the meta data portion for each response.
type Meta struct {
	Result  int    `json:"result"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
	Command string `json:"command"`
}

// New returns a plat from given env.
func New(env *env.Env) *Plat {
	return &Plat{
		env: env,

		acters: []acter.Acter{
			act.NewExiler(env),
			act.NewQuarantiner(env),
			act.NewCleaner(env),
			act.NewAlerter(env),
		},

		cfg: &Cfg{
			path:     filepath.Join(env.Paths.Plat.Dir, "cpanel.toml"),
			SkipAccs: []string{""},
		},

		// Path to whmapi. Must be absolute for CL compat https://api.docs.cpanel.net/whm/introduction
		bin: "/usr/local/cpanel/bin/apitool",

		domainInfo: []string{"get_domain_info"},
	}
}

// Load reads given plat cfg files.
func (plat *Plat) Load() error {
	enabled := []acter.Acter{}

	for _, acter := range plat.acters {
		if err := acter.Load(); err == nil {
			enabled = append(enabled, acter)
		}
	}

	plat.acters = enabled

	re.SetTargets(reTarget)

	tmps := []string{
		"/tmp",
		"/var/tmp",
		"/dev/shm",
	}

	paths, err := plat.GetDocRoots()
	if err != nil {
		return err
	}

	for _, path := range append(paths, tmps...) {
		if slices.Contains(plat.cfg.SkipAccs, re.Target(path)) ||
			slices.Contains(plat.env.Cfg.Scans.Paths, path) {
			continue
		}

		plat.env.Cfg.Scans.Paths = append(plat.env.Cfg.Scans.Paths, path)
	}

	return nil
}

// GetDocRoots performs a get_domain_info request and returns document root paths.
func (plat *Plat) GetDocRoots() ([]string, error) {
	var (
		info  = &DomainInfo{}
		args  = append(plat.domainInfo, "--output=jsonpretty")
		paths []string
	)

	// Tests override.
	if plat.bin == "echo" {
		args = plat.domainInfo
	}

	resp, err := plat.Exec(args...)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resp, info); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrAPIDomInfoUnmarshal, err)
	}

	for _, path := range info.Data.Domains {
		paths = append(paths, path.Docroot, filepath.Join(filepath.Dir(path.Docroot), "tmp"))
	}

	return paths, nil
}

// Exec performs an api lookup using whmapi cmd.
func (plat *Plat) Exec(args ...string) ([]byte, error) {
	buff := bytes.Buffer{}

	cmd := exec.Command(plat.bin, args...)
	cmd.Stdout = &buff

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrAPIToolExec, err)
	}

	return buff.Bytes(), nil
}

// Acters returns a given cpanel plat's active acts.
func (plat *Plat) Acters() []acter.Acter {
	return plat.acters
}

// Cfg returns a given cpanel plat's cfg.
func (plat *Plat) Cfg() plat.Cfg {
	return plat.cfg
}
