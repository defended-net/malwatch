// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/exec"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

// Plat represents a directadmin platform.
type Plat struct {
	env       *env.Env
	cfg       *Cfg
	acters    []acter.Acter
	bin       string
	url       string
	endpoints *endpoints
}

// Info represents top level api response.
type Info struct {
	Users map[string]users
}

// users represents user:domains.
type users map[string]domains

// domains represents domain:docroots.
type domains map[string]*docroot

type docroot struct {
	PrivateHTML string  `json:"private_html"`
	PublicHTML  string  `json:"public_html"`
	Subdomains  domains `json:"subdomains"`
}

type endpoints struct {
	docroots *endpoint
}

type endpoint struct {
	url    string
	base   string
	params map[string]string
}

// New returns a plat from given env.
func New(env *env.Env) *Plat {
	return &Plat{
		env:    env,
		acters: act.Preset(env),

		cfg: &Cfg{
			path: filepath.Join(env.Paths.Plat.Dir, "directadmin.toml"),
		},

		// https://docs.directadmin.com/directadmin/general-usage/directadmin-binary.html
		bin: "/usr/local/directadmin/directadmin",

		endpoints: &endpoints{
			docroots: &endpoint{
				base: "/CMD_API_DOMAIN",

				params: map[string]string{
					"action": "document_root_all",
				},
			},
		},
	}
}

// Load reads given plat's cfgs.
func (plat *Plat) Load() error {
	if err := plat.Auth(plat.cfg.User); err != nil {
		return err
	}

	acters, err := acter.Load(plat.acters)
	if err != nil {
		return err
	}

	plat.acters = acters

	for _, endpoint := range []*endpoint{
		plat.endpoints.docroots,
	} {
		if err := endpoint.Prep(plat.url); err != nil {
			return err
		}
	}

	re.SetTargets(reTarget)

	tmps := []string{
		"/tmp",
		"/var/tmp",
		"/dev/shm",
	}

	paths, err := plat.DocRoots()
	if err != nil {
		return err
	}

	for _, path := range append(paths, tmps...) {
		if !slices.Contains(plat.cfg.SkipAccs, re.Target(path)) &&
			!slices.Contains(plat.env.Cfg.Scans.Paths, path) {
			plat.env.Cfg.Scans.Paths = append(plat.env.Cfg.Scans.Paths, path)
		}
	}

	return nil
}

// DocRoots performs a get_domain_info req and returns document root paths.
func (plat *Plat) DocRoots() ([]string, error) {
	var (
		info  = &Info{}
		dedup = map[string]struct{}{}
		paths []string
	)

	resp, err := http.Get(plat.endpoints.docroots.url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, info); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrAPIDomInfoUnmarshal, err)
	}

	for _, domains := range info.Users {
		for _, metas := range domains {
			for _, meta := range metas {
				dedup[filepath.Join(filepath.Dir(meta.PublicHTML), "tmp")] = struct{}{}

				dedup[meta.PublicHTML] = struct{}{}
				dedup[meta.PrivateHTML] = struct{}{}

				for _, subdomain := range meta.Subdomains {
					if fsys.IsRel(subdomain.PublicHTML, meta.PublicHTML) {
						continue
					}

					// private_html has same base dir so can bundle together.
					dedup[subdomain.PublicHTML] = struct{}{}
					dedup[subdomain.PrivateHTML] = struct{}{}
				}
			}
		}
	}

	for path := range dedup {
		if path != "" {
			paths = append(paths, path)
		}
	}

	slices.Sort(paths)

	return paths, nil
}

// Auth auths for given user.
func (plat *Plat) Auth(user string) error {
	if plat.url != "" {
		return nil
	}

	url, err := exec.Run(plat.bin, "root-auth-url", fmt.Sprintf("--user=%s", user))
	if err != nil {
		return fmt.Errorf("%w, %v", ErrAPIExec, err)
	}

	plat.url = strings.TrimSuffix(string(url), "\n")

	return nil
}

// Prep prepares given api endpoint by appending the base followed by adding params.
// Encoded url is then stored.
func (endpoint *endpoint) Prep(authURL string) error {
	path, err := url.JoinPath(authURL, endpoint.base)
	if err != nil {
		return err
	}

	params := url.Values{}

	for name, val := range endpoint.params {
		params.Add(name, val)
	}

	endpoint.url = path + "?" + params.Encode()

	return nil
}

// Acters returns given plat's enabled acts.
func (plat *Plat) Acters() []acter.Acter {
	return plat.acters
}

// Cfg returns given plat's cfg.
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

		endpoints: &endpoints{
			docroots: &endpoint{
				base: "/CMD_API_DOMAIN",

				params: map[string]string{
					"action": "document_root_all",
				},
			},
		},
	}, nil
}
