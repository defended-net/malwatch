// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package git

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

const repo = "https://github.com/defended-net/malwatch-signatures"

func TestClone(t *testing.T) {
	repo := &secret.Repo{
		URL: repo,
	}

	tag, err := Clone(repo, t.TempDir())
	if err != nil {
		t.Fatalf("clone error: %s", err)
	}

	fmt.Println(tag)
}

func TestLatestTag(t *testing.T) {
	if _, err := http.Get(repo); err != nil {
		t.Errorf("http get error: %s", err)
	}
}
