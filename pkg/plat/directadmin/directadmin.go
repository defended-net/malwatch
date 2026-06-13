// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

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
	client    *http.Client
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

		client: &http.Client{
			Timeout: 10 * time.Second,
		},

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
func (plat *Plat) Load(root *os.Root) error {
	if err := plat.Auth(plat.cfg.User); err != nil {
		return err
	}

	acters, err := acter.Load(root, plat.acters)
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
		info   = &Info{}
		dedup  = map[string]struct{}{}
		paths  []string
		bodySz = int64(10 << 20)
	)

	client := plat.client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get(plat.endpoints.docroots.url)
	if err != nil {
		return nil, err
	}
	defer fsys.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w, %v", ErrAPIRespCode, resp.StatusCode)
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, bodySz)).Decode(info); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrAPIDomInfoUnmarshal, err)
	}

	for _, domains := range info.Users {
		for _, metas := range domains {
			for _, meta := range metas {
				if !addDocroot(dedup, meta.PublicHTML, meta.PrivateHTML) {
					continue
				}

				dedup[filepath.Join(filepath.Dir(meta.PublicHTML), "tmp")] = struct{}{}

				for _, subdomain := range meta.Subdomains {
					if fsys.IsRel(subdomain.PublicHTML, meta.PublicHTML) {
						continue
					}

					addDocroot(dedup, subdomain.PublicHTML, subdomain.PrivateHTML)
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

	if !reUser.MatchString(user) {
		return fmt.Errorf("%w, %v", ErrAPIAuthUser, user)
	}

	presigned, err := exec.Run(plat.bin, "root-auth-url", fmt.Sprintf("--user=%s", user))
	if err != nil {
		return fmt.Errorf("%w, %v", ErrAPIExec, err)
	}

	parsed, err := url.Parse(strings.TrimSuffix(string(presigned), "\n"))
	if err != nil {
		return fmt.Errorf("%w, %v", ErrAPIAuthURL, err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("%w, got %q", ErrAPIAuthScheme, parsed.Scheme)
	}

	var (
		allowed = []string{"localhost", "127.0.0.1", "::1"}
		host    = parsed.Hostname()
	)

	if !slices.Contains(allowed, host) {
		return fmt.Errorf("%w, got %q", ErrAPIAuthHost, host)
	}

	plat.url = parsed.String()

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

// addDocroot validates and adds paths for given dedup.
func addDocroot(dedup map[string]struct{}, paths ...string) bool {
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			slog.Info(fsys.ErrPathNotAbs.Error(), "path", path)

			return false
		}
	}

	if err := fsys.HasDotDots(paths...); err != nil {
		slog.Info(fsys.ErrPathTraverse.Error(), "paths", paths)

		return false
	}

	for _, path := range paths {
		dedup[path] = struct{}{}
	}

	return true
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
			User: "user",
		},

		bin: "echo",
		url: "",

		client: &http.Client{
			Timeout: 10 * time.Second,
		},

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
