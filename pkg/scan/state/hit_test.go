// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"io"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/defended-net/malwatch/pkg/db/orm/hit"
)

func TestNewJob(t *testing.T) {
	var (
		want = &Job{
			WGrp: &sync.WaitGroup{},
			Hits: make(chan *Hit),
			errs: &Errs{},
		}

		got = NewJob()
	)

	got.Hits = want.Hits

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected new job result %v, want %v", got, want)
	}
}

func TestGroup(t *testing.T) {
	var (
		tmp = t.TempDir()

		got = Group("fs",
			[]*Hit{
				{
					Path: tmp,
					Meta: &hit.Meta{},
				},
			},
		)

		want = []*Result{
			NewResult("fs",
				Paths{
					tmp: &hit.Meta{},
				},
			),
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected group result %v, want %v", got, want)
	}
}

func TestGroupEmpty(t *testing.T) {
	var (
		tmp  = t.TempDir()
		meta = &hit.Meta{}

		hits = []*Hit{
			{
				Path: tmp, Meta: meta,
			},
		}

		got = Group("", hits)
	)

	if got[0].Target == "" {
		t.Errorf("expected target to be set, got empty target")
	}
}

func TestGroupDupe(t *testing.T) {
	var (
		tmp = t.TempDir()

		meta = &hit.Meta{
			Rules: []string{
				"rule-one",
				"rule-two",
			},
		}

		want = []*Result{
			{
				Target: "fs",

				Paths: Paths{
					tmp: &hit.Meta{
						Rules: []string{
							"rule-one",
							"rule-two",
							"rule-three",
						},
					},
				},

				errs: &Errs{},
			},
		}

		hits = []*Hit{
			{
				Path: tmp,
				Meta: meta,
			},

			{
				Path: tmp,

				Meta: &hit.Meta{
					Rules: []string{
						"rule-three",
					},
				},
			},
		}

		got = Group("fs", hits)
	)

	if !slices.Equal(got[0].Paths[tmp].Rules, want[0].Paths[tmp].Rules) {
		t.Errorf("unexpected group result %v, want %v", got, want)
	}
}

func TestGroupCompound(t *testing.T) {
	var (
		dir1, dir2 = t.TempDir(), t.TempDir()

		want = []*Result{
			{
				Target: "fs",

				Paths: Paths{
					dir1: &hit.Meta{},
					dir2: &hit.Meta{},
				},

				errs: &Errs{},
			},
		}

		hits = []*Hit{
			{
				Path: dir1,
				Meta: &hit.Meta{},
			},

			{
				Path: dir2,
				Meta: &hit.Meta{},
			},
		}

		got = Group("fs", hits)
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected group result %v, want %v", got, want)
	}
}

func TestAddErr(t *testing.T) {
	tests := map[string]struct {
		job   *Job
		path  string
		input []error
		want  []error
	}{
		"new": {
			job: &Job{
				errs: &Errs{},
			},

			input: []error{
				io.EOF,
			},

			want: []error{
				io.EOF,
			},
		},

		"append": {
			job: &Job{
				errs: &Errs{
					Vals: []error{
						io.EOF,
					},
				},
			},

			input: []error{
				io.EOF,
			},

			want: []error{
				io.EOF,
				io.EOF,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, err := range test.input {
				test.job.AddErr(err)
			}

			if got := test.job.Errs(); !slices.Equal(got, test.want) {
				t.Errorf("unexpected get errs result %v, want %v", got, test.want)
			}
		})
	}
}
