// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import (
	"reflect"
	"slices"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

var mock = &Plat{
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
                "parent_domain": "first.com",
                "docroot": "/home/first/public_html",
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
                "docroot": "/home/second/public_html",
                "parent_domain": "second.com",
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
                "parent_domain": "third.com",
                "docroot": "/home/third/public_html",
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
}

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Paths.Plat = &path.Plat{
		Dir: t.TempDir(),
	}

	plat := New(env)

	plat.bin = "echo"
	plat.domainInfo = mock.domainInfo

	if err = plat.Load(); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestExec(t *testing.T) {
	mock := &Plat{
		bin: "echo",
	}

	result, err := mock.Exec(t.Name())
	if err != nil {
		t.Fatalf("exec error: %v", err)
	}

	if string(result) != t.Name()+"\n" {
		t.Errorf("unexpected exec result %v, want %v", string(result), t.Name())
	}
}

func TestGetDomainInfo(t *testing.T) {
	want := []string{
		"/home/first/public_html",
		"/home/first/tmp",
		"/home/second/public_html",
		"/home/second/tmp",
		"/home/third/public_html",
		"/home/third/tmp",
	}

	result, err := mock.GetDocRoots()
	if err != nil {
		t.Fatalf("get domain error: %v", err)
	}

	if !slices.Equal(result, want) {
		t.Errorf("unexpected get domain info result %v, want %v", result, want)
	}
}

func TestGetDomainInfoErrs(t *testing.T) {
	mock := &Plat{
		bin: t.Name(),
	}

	if _, err := mock.GetDocRoots(); err == nil {
		t.Errorf("unexpected get domain success")
	}
}

func TestCfg(t *testing.T) {
	var (
		plat = &Plat{
			cfg: &Cfg{},
		}

		got = plat.Cfg()
	)

	if !reflect.DeepEqual(got, plat.cfg) {
		t.Errorf("unexpected cfg result %v, want %v", got, plat.cfg)
	}
}

func TestActers(t *testing.T) {
	var (
		input = acter.Mock(act.VerbAlert)

		plat = &Plat{
			acters: []acter.Acter{
				input,
			},
		}

		got = plat.Acters()
	)

	if !reflect.DeepEqual(got, plat.acters) {
		t.Errorf("unexpected acts result %v, want %v", got, plat.acters)
	}
}
