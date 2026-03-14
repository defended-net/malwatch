// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"slices"

	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
)

// Hit represents a hit.
type Hit struct {
	Path string
	Meta *hit.Meta
}

// Paths represents hit meta per path.
type Paths map[string]*hit.Meta

// Group returns a slice of results from given slice of hits.
func Group(target string, hits []*Hit) []*Result {
	if len(hits) == 0 {
		return []*Result{NewResult(target, Paths{})}
	}

	grouped := make(map[string]*Result, len(hits))

	for _, hit := range hits {
		target := re.Target(hit.Path)

		add(grouped, target, hit)
	}

	results := make([]*Result, 0, len(grouped))

	for _, result := range grouped {
		results = append(results, result)
	}

	return results
}

func add(grouped map[string]*Result, target string, hit *Hit) {
	result := grouped[target]

	if result == nil {
		grouped[target] = NewResult(
			target,

			Paths{
				hit.Path: hit.Meta,
			},
		)

		return
	}

	current := result.Paths[hit.Path]

	if current == nil {
		result.Paths[hit.Path] = hit.Meta

		return
	}

	merge(hit.Meta, current)
}

func merge(src *hit.Meta, dst *hit.Meta) {
	for _, rule := range src.Rules {
		if !slices.Contains(dst.Rules, rule) {
			dst.Rules = append(dst.Rules, rule)
		}
	}
}
