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

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		input := strings.ToLower(scanner.Text())

		switch input {
		case "y":
			return true

		case "n":
			return false
		}

		fmt.Printf("%s [y/n] ", msg)
	}

	if err := scanner.Err(); err != nil {
		slog.Error(err.Error())
	}

	return false
}
