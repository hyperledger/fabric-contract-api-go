# Support and Compatibility for fabric-contract-api-go

GitHub is used for code base management, issues should be reported in the [FABCAG](https://jira.hyperledger.org/projects/FABCAG/issues) component in JIRA.

## Summary of Compatibility

This table shows the summary of the compatibility of the package at version 1.0, together with the Go version it requires and the Fabric Peer version it can communicate with.

|            | Tested | Supported |
| ---------- | ------ | --------- |
| Fabric     | 2.1    | 2.0.x     |
| Go         | 1.13   | 1.13+     |

By default a Fabric Peer v2.0 will produce a chaincode docker image using Fabric ccenv v2.0, this uses Go 1.13.

## Compatibility

The key elements are:
- The version of fabric-contract-api-go packages used
- The version of Go used to build the code
- When starting a chaincode container to run a Smart Contract the version of Go used to build the Smart Contract is determined by these factors:

Fabric v2.0+ will, by default, start up a docker image to host the chaincode and contracts. The version of Go used is therefore determined by Fabric.

With Fabric v2.0+, the chaincode container can be configured to be started by other means, and not the Peer. In this case, the version of Go used to build the container is not in the control of Fabric.

## Supported Go versions

v1.0.0 packages are supported for building with Go 1.13.