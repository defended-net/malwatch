// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package batch

import (
	"slices"
	"testing"

	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestHitGet(t *testing.T) {
	var (
		want = &state.Hit{
			Path: t.TempDir(),
			Meta: &hit.Meta{},
		}

		hits = &Hits{}
	)

	hits.Add(want)

	got := hits.Get(false)

	if !slices.Equal(got, []*state.Hit{want}) {
		t.Errorf("unexpected hits get result %v, want %v", got, want)
	}
}

func TestHitAdd(t *testing.T) {
	var (
		want = &state.Hit{
			Path: t.TempDir(),
			Meta: &hit.Meta{},
		}

		hits = &Hits{}
	)

	hits.Add(want)

	got := hits.Get(false)

	if !slices.Equal(got, []*state.Hit{want}) {
		t.Errorf("unexpected hits add result %v, want %v", got, want)
	}
}
