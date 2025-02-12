// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/tui/tbl"
)

// Print prints given result's hits.
func (result *Result) Print() error {
	// Only print if malwatch scan /path
	if len(os.Args) < 3 {
		return nil
	}

	rows := [][]string{}

	for path, meta := range result.Paths {
		rows = append(rows, []string{
			path,
			strings.Join(meta.Rules, "\n"),
			meta.Status,
			fmt.Sprint(meta.Acts),
		})
	}

	return tbl.Print(result.Target, tbl.HdrFileReport, rows)
}

// Save saves given result's hits to given db.
func (result *Result) Save(db *bbolt.DB) error {
	if db == nil || len(result.Paths) == 0 {
		return nil
	}

	history := &hit.History{
		Target: result.Target,
		Paths:  hit.Paths{},
	}

	for path, meta := range result.Paths {
		history.Paths[path] = []*hit.Meta{meta}
	}

	return history.Save(db)
}

// Log logs given result's hits.
func (result *Result) Log() error {
	slog.Info("scan", "target", result.Target, "hits", result.Paths)

	return nil
}
