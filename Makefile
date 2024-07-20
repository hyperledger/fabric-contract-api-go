# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
functional_dir := $(base_dir)/internal/functionaltests
go_bin_dir := $(shell go env GOPATH)/bin

mockery_version := 2.43.2
kernel_name := $(shell uname -s)
machine_hardware := $(shell uname -m)

.PHONY: test
test: generate lint unit-test functional-test

.PHONY: lint
lint: staticcheck golangci-lint

.PHONY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	cd '$(base_dir)' && staticcheck -f stylish ./...

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b '$(go_bin_dir)'

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	cd '$(base_dir)' && golangci-lint run

.PHONY: install-mockery
install-mockery:
	curl --fail --location \
		'https://github.com/vektra/mockery/releases/download/v$(mockery_version)/mockery_$(mockery_version)_$(kernel_name)_$(machine_hardware).tar.gz' \
		| tar -C '$(go_bin_dir)' -xzf - mockery

$(go_bin_dir)/mockery:
	$(MAKE) install-mockery

.PHONY: generate
generate: $(go_bin_dir)/mockery
	cd '$(base_dir)' && mockery

.PHONY: unit-test
unit-test:
	cd '$(base_dir)' && go test -race $$(go list ./... | grep -v functionaltests)

.PHONY: functional-test
functional-test:
	go install github.com/cucumber/godog/cmd/godog@v0.12
	cd '$(functional_dir)' && godog run features/*

.PHONY: scan
scan:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	cd '$(base_dir)' && govulncheck ./...
