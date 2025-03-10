// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"slices"

	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
)

// Hit represents a hit detection.
type Hit struct {
	Path string
	Meta *hit.Meta
}

// Paths represents hit meta per path.
type Paths map[string]*hit.Meta

// Group returns a slice of results from given slice of hits.
func Group(target string, hits []*Hit) []*Result {
	if len(hits) == 0 {
		return []*Result{
			NewResult(target, Paths{}),
		}
	}

	grouped := map[string]*Result{}

	for _, hit := range hits {
		target = re.Target(hit.Path)

		result, exists := grouped[target]
		if !exists {
			grouped[target] = NewResult(
				target,

				Paths{
					hit.Path: hit.Meta,
				})

			continue
		}

		meta, exists := result.Paths[hit.Path]
		if !exists {
			result.Paths[hit.Path] = hit.Meta
			continue
		}

		for _, rule := range hit.Meta.Rules {
			if !slices.Contains(meta.Rules, rule) {
				meta.Rules = append(meta.Rules, rule)
			}
		}
	}

	results := make([]*Result, 0, len(grouped))

	for _, result := range grouped {
		results = append(results, result)
	}

	return results
}
