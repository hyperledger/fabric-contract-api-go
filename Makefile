# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
functional_dir := $(base_dir)/internal/functionaltests
go_bin_dir := $(shell go env GOPATH)/bin

mockery := $(go_bin_dir)/mockery
osv_scanner := $(go_bin_dir)/osv-scanner
golangci_lint := $(go_bin_dir)/golangci-lint

mockery_version := 3.7.4

kernel_name := $(shell uname -s)
lowercase_kernel_name := $(shell echo '$(kernel_name)' | tr '[:upper:]' '[:lower:]')

machine_hardware := $(shell uname -m)
ifeq ($(machine_hardware), aarch64)
	machine_hardware := arm64
endif

amd_arm_machine_hardware := $(machine_hardware)
ifeq ($(machine_hardware), x86_64)
	amd_arm_machine_hardware := amd64
endif

TMPDIR ?= /tmp
TMPDIR := $(abspath $(TMPDIR))

.PHONY: test
test: generate lint unit-test functional-test

.PHONY: lint
lint: golangci-lint

.PHONY: install-golangci-lint
install-golangci-lint: uninstall-golangci-lint $(golangci_lint)

.PHONY: uninstall-golangci-lint
uninstall-golangci-lint:
	rm -f '$(golangci_lint)'

$(golangci_lint):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b '$(go_bin_dir)'

.PHONY: golangci-lint
golangci-lint: $(golangci_lint)
	cd '$(base_dir)' && '$(golangci_lint)' run

.PHONY: install-mockery
install-mockery: uninstall-mockery $(mockery)

.PHONY: uninstall-mockery
uninstall-mockery:
	rm -f '$(mockery)'

$(mockery):
	curl --fail --location \
		'https://github.com/vektra/mockery/releases/download/v$(mockery_version)/mockery_$(mockery_version)_$(kernel_name)_$(machine_hardware).tar.gz' \
		| tar -C '$(go_bin_dir)' -xzf - mockery
	chmod u+x '$(mockery)'

.PHONY: generate
generate: $(mockery)
	cd '$(base_dir)' && '$(mockery)'

.PHONY: unit-test
unit-test:
	cd '$(base_dir)' && go test -race $$(go list ./... | grep -v functionaltests)

.PHONY: functional-test
functional-test:
	cd '$(functional_dir)' && go test -test.run '^TestFeatures$$'

.PHONY: install-osv-scanner
install-osv-scanner: uninstall-osv-scanner $(osv_scanner)

.PHONY: uninstall-osv-scanner
uninstall-osv-scanner:
	rm -f '$(osv_scanner)'

$(osv_scanner):
	curl --fail --location --show-error --silent --output '$(osv_scanner)' \
    	'https://github.com/google/osv-scanner/releases/latest/download/osv-scanner_$(lowercase_kernel_name)_$(amd_arm_machine_hardware)'
	chmod u+x '$(osv_scanner)'

.PHONY: scan
scan: $(osv_scanner)
	echo "GoVersionOverride = '$$(go env GOVERSION | sed -e 's/^go//' -e 's/-.*//')'" > '$(TMPDIR)/osv-scanner.toml'
	'$(osv_scanner)' scan source --config='$(TMPDIR)/osv-scanner.toml' --lockfile='$(base_dir)/go.mod'

PHONY: sync-deps
sync-deps:
	cd '$(base_dir)' && go mod tidy \
		&& cd '$(base_dir)/integrationtest/chaincode' \
		&& find . -mindepth 2 -maxdepth 2 -type f -name go.mod -execdir go mod tidy \;
