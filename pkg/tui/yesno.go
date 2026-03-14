// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// YesNo displays a y/n prompt with given msg.
func YesNo(msg string, reader io.Reader) bool {
	fmt.Printf("%s [y/n] ", msg)

	var (
		scanner = bufio.NewScanner(reader)
		input   = ""
	)

	for scanner.Scan() {
		input = strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch input {
		case "y", "yes":
			return true

		case "n", "no":
			return false
		}

		fmt.Printf("%s [y/n] ", msg)
	}

	if err := scanner.Err(); err != nil {
		slog.Error(err.Error())
	}

	return false
}
