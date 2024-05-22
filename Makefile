# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
functional_dir := $(base_dir)/internal/functionaltests
go_bin_dir := $(shell go env GOPATH)/bin

.PHONY: lint
lint: staticcheck golangci-lint

.PHONY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck -f stylish './...'

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b '$(go_bin_dir)'

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONY: unit-test
unit-test:
	go test -race $$(go list '$(base_dir)/...' | grep -v functionaltests)

.PHONY: functional-test
functional-test:
	go install github.com/cucumber/godog/cmd/godog@v0.12
	cd '$(functional_dir)' && godog run features/*

.PHONY: scan
scan:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck '$(base_dir)/...'
