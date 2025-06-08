// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/exec"
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
		env:    env,
		acters: act.Preset(env),

		cfg: &Cfg{
			path: filepath.Join(env.Paths.Plat.Dir, "cpanel.toml"),
		},

		// Path to whmapi. Must be absolute for CL compat https://api.docs.cpanel.net/whm/introduction
		bin: "/usr/local/cpanel/bin/apitool",

		domainInfo: []string{"get_domain_info"},
	}
}

// Load reads given plat's cfgs.
func (plat *Plat) Load() error {
	acters, err := acter.Load(plat.acters)
	if err != nil {
		return err
	}

	plat.acters = acters

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

// GetDocRoots performs a get_domain_info request and returns docroot paths.
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

	resp, err := exec.Run(plat.bin, args...)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(resp), info); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrAPIDomInfoUnmarshal, err)
	}

	for _, path := range info.Data.Domains {
		paths = append(paths, path.Docroot, filepath.Join(filepath.Dir(path.Docroot), "tmp"))
	}

	return paths, nil
}

// Acters returns a given cpanel plat's active acts.
func (plat *Plat) Acters() []acter.Acter {
	return plat.acters
}

// Cfg returns a given cpanel plat's cfg.
func (plat *Plat) Cfg() plat.Cfg {
	return plat.cfg
}

// Mock mocks a plat.
func Mock(name string, dir string) (*Plat, error) {
	env, err := env.Mock(name, dir)
	if err != nil {
		return nil, err
	}

	return &Plat{
		env: env,

		acters: []acter.Acter{
			acter.Mock(name, true),
		},

		cfg: &Cfg{
			path: filepath.Join(env.Paths.Plat.Dir, "directadmin.toml"),
		},

		bin: "echo",

		domainInfo: []string{`{
    "data": {
        "domains": [
            {
                "ipv4": "123.123.123.123",
                "modsecurity_enabled": 1,
                "php_version": "ea-php81",
                "user_owner": "first",
                "user": "first",
                "domain_type": "main",
                "ipv6": null,
                "port": "80",
                "port_ssl": "443",
                "ipv6_is_dedicated": 0,
                "domain": "first.com",
                "parent_domain": "one.example",
                "docroot": "/home/one/public_html",
                "ipv4_ssl": "123.123.123.123"
            },

            {
                "ipv6": null,
                "port": "80",
                "php_version": "ea-php81",
                "modsecurity_enabled": 1,
                "ipv4": "123.123.123.123",
                "user": "second",
                "domain_type": "main",
                "user_owner": "second",
                "ipv4_ssl": "123.123.123.123",
                "docroot": "/home/two/public_html",
                "parent_domain": "two.example",
                "port_ssl": "443",
                "domain": "second.com",
                "ipv6_is_dedicated": 0
            },

            {
                "ipv4": "123.123.123.123",
                "modsecurity_enabled": 1,
                "php_version": "ea-php81",
                "user_owner": "third",
                "user": "third",
                "domain_type": "main",
                "ipv6": null,
                "port": "80",
                "port_ssl": "443",
                "ipv6_is_dedicated": 0,
                "domain": "third.com",
                "parent_domain": "three.example",
                "docroot": "/home/three/public_html",
                "ipv4_ssl": "123.123.123.123"
            }
        ]
    },
    
    "metadata": {
        "result": 1,
        "version": 1,
        "reason": "OK",
        "command": "get_domain_info"
    }
}`,
		},
	}, nil
}
