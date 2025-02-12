// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tbl

import (
	"fmt"

	"github.com/alexeyco/simpletable"
)

var (
	// HdrAppHit stores the header for app hits.
	HdrAppHit = []string{"Path", "Date", "Malware", "Created", "Modified", "User", "Group", "Status", "Actions"}

	// HdrFileReport stores the header for file report.
	HdrFileReport = []string{"Path", "Malware", "Status", "Actions"}
)

// Prepare formats header and row data for table output.
func Prepare(hdrs []string, rows [][]string) ([]string, [][]string) {
	for rIdx, row := range rows {
		rows[rIdx] = make([]string, len(hdrs))

		for cIdx := range row {
			switch {
			case cIdx < len(hdrs):
				rows[rIdx][cIdx] = row[cIdx]

			case cIdx%(len(hdrs)) == 0:
				rows[rIdx][0] += "\n"

			default:
				rows[rIdx][cIdx%len(hdrs)] = rows[rIdx][cIdx%len(hdrs)] + "\n" + row[cIdx]
			}
		}
	}

	return hdrs, rows
}

// Print prints a table.
func Print(title string, hdrs []string, rows [][]string) error {
	tbl := simpletable.New()
	tbl.SetStyle(simpletable.StyleUnicode)

	hdrs, rows = Prepare(hdrs, rows)

	for _, header := range hdrs {
		cell := &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  header,
		}

		tbl.Header.Cells = append(tbl.Header.Cells, cell)
	}

	tbl.Header.Cells[0].Text = fmt.Sprintf("[%v] %v", title, tbl.Header.Cells[0].Text)

	for _, row := range rows {
		cells := make([]*simpletable.Cell, len(row))

		for cIdx, cell := range row {
			cells[cIdx] = &simpletable.Cell{Text: cell}
		}

		tbl.Body.Cells = append(tbl.Body.Cells, cells)
	}

	tbl.Println()

	return nil
}
