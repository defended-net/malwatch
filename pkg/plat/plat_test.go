// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"slices"
	"testing"

	"github.com/defended-net/malwatch/pkg/plat/acter"
)

func TestLoad(t *testing.T) {
	mock := Mock(acter.Mock(t.Name()))

	if err := mock.Load(); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestCfg(t *testing.T) {
	var (
		mock = &mock{}
		want = cfg{}
	)

	if got := mock.Cfg(); got != want {
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}
}

func TestMock(t *testing.T) {
	var (
		want = []acter.Acter{
			acter.Mock(t.Name()),
		}

		got = Mock(want...)
	)

	if !slices.Equal(got.Acters(), want) {
		t.Errorf("unexpected mock acters %v, want %v", got.Acters(), want)
	}
}
