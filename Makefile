# © Roscoe Skeens <rskeens@defended.net>
# SPDX-License-Identifier: AGPL-3.0-or-later

## help: print help.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## audit: verify modules, vet and static check. Then scan for vulnerabilities.
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: test all pkgs.
.PHONY: test
test:
	go test ./...

## race: test with race condition checks.
.PHONY: race
race:
	go test -race ./...

## build: compile binary to dir /tmp/malwatch/
.PHONY: build
build:
	go build -trimpath -ldflags="-w -s" -o=/tmp/malwatch/ ./cmd/malwatch
