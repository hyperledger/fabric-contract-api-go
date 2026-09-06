# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

BASE_DIR := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
FUNCTIONAL_DIR := $(BASE_DIR)/internal/functionaltests
GO_BIN_DIR := $(shell go env GOBIN)
ifeq ($(GO_BIN_DIR),)
	GO_BIN_DIR := $(shell go env GOPATH)/bin
endif

MOCKERY := $(GO_BIN_DIR)/mockery
OSV_SCANNER := $(GO_BIN_DIR)/osv-scanner
GOLANGCI_LINT := $(GO_BIN_DIR)/golangci-lint

KERNEL_NAME := $(shell uname -s)
LOWERCASE_KERNEL_NAME := $(shell echo '$(KERNEL_NAME)' | tr '[:upper:]' '[:lower:]')

MACHINE_HARDWARE := $(shell uname -m)
ifeq ($(MACHINE_HARDWARE), aarch64)
	MACHINE_HARDWARE := arm64
endif

AMD_ARM_MACHINE_HARDWARE := $(MACHINE_HARDWARE)
ifeq ($(AMD_ARM_MACHINE_HARDWARE), x86_64)
	AMD_ARM_MACHINE_HARDWARE := amd64
endif

TMPDIR ?= /tmp
TMPDIR := $(abspath $(TMPDIR))

# If GH_TOKEN environment variable is set, use it as the GitHub API auth token
GH_API_AUTH := $(if $(GH_TOKEN),--header 'Authorization: Bearer $(GH_TOKEN)',)

.PHONY: test
test: lint unit-test functional-test

.PHONY: lint
lint: generate golangci-lint

.PHONY: install-golangci-lint
install-golangci-lint: uninstall-golangci-lint $(GOLANGCI_LINT)

.PHONY: uninstall-golangci-lint
uninstall-golangci-lint:
	rm -f '$(GOLANGCI_LINT)'

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b '$(dir $(GOLANGCI_LINT))'

.PHONY: golangci-lint
golangci-lint: generate $(GOLANGCI_LINT)
	cd '$(BASE_DIR)' && '$(GOLANGCI_LINT)' run

.PHONY: install-mockery
install-mockery: uninstall-mockery $(MOCKERY)

.PHONY: uninstall-mockery
uninstall-mockery:
	rm -f '$(MOCKERY)'

# Silent to prevent printing of auth token
.SILENT: $(MOCKERY)
$(MOCKERY):
	mockery_version=$$(curl --fail --show-error --silent $(GH_API_AUTH) https://api.github.com/repos/vektra/mockery/releases | jq --raw-output '.[].tag_name' | sort --version-sort | tail -1) && \
		curl --fail --location --show-error --silent \
			"https://github.com/vektra/mockery/releases/download/$${mockery_version}/mockery_$${mockery_version#v}_$(KERNEL_NAME)_$(MACHINE_HARDWARE).tar.gz" \
			| tar -C '$(dir $(MOCKERY))' -xzf - mockery
	chmod u+x '$(MOCKERY)'

.PHONY: generate
generate: contractapi/mocks_test.go

contractapi/mocks_test.go: $(MOCKERY)
	cd '$(BASE_DIR)' && '$(MOCKERY)'

.PHONY: unit-test
unit-test: generate
	cd '$(BASE_DIR)' && go test -race $$(go list ./... | grep -v functionaltests)

.PHONY: functional-test
functional-test:
	cd '$(FUNCTIONAL_DIR)' && go test -test.run '^TestFeatures$$'

.PHONY: install-osv-scanner
install-osv-scanner: uninstall-osv-scanner $(OSV_SCANNER)

.PHONY: uninstall-osv-scanner
uninstall-osv-scanner:
	rm -f '$(OSV_SCANNER)'

$(OSV_SCANNER):
	curl --fail --location --show-error --silent --output '$(OSV_SCANNER)' \
    	'https://github.com/google/osv-scanner/releases/latest/download/osv-scanner_$(LOWERCASE_KERNEL_NAME)_$(AMD_ARM_MACHINE_HARDWARE)'
	chmod u+x '$(OSV_SCANNER)'

.PHONY: scan
scan: $(OSV_SCANNER)
	echo "GoVersionOverride = '$$(go env GOVERSION | sed -e 's/^go//' -e 's/-.*//')'" > '$(TMPDIR)/osv-scanner.toml'
	'$(OSV_SCANNER)' scan source --config='$(TMPDIR)/osv-scanner.toml' --lockfile='$(BASE_DIR)/go.mod'

.PHONY: sync-deps
sync-deps:
	cd '$(BASE_DIR)' && go mod tidy \
		&& cd '$(BASE_DIR)/integrationtest/chaincode' \
		&& find . -mindepth 2 -maxdepth 2 -type f -name go.mod -execdir go mod tidy \;
