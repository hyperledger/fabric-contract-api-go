# Contributing

We welcome contributions to the Hyperledger Fabric Project in many forms, and there's always plenty to do!

Please visit the [contributors guide](https://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING.html) in the docs to learn how to make contributions to this exciting project.

## Folder Structure
This repo contains multiple packages that are externalised to users of the contract API: contractapi, metadata and serializer. The internal folder contains contents which are not designed to be used by developers consuming the packages provided by this repo, but which are used by other packages in the repo. 

Unit tests for each package are located within the package and follow the pattern of `<FILE_TESTED>_test.go`.

There is a single go.mod file for handling the Go modules of all of the packages within this repo.

## Developing for this repo
Although unit test coverage of 100% is not always possible or sensible in Go, unit tests should nonetheless be written for as much of your code as necessary. Test coverage is checked during the testing process. Should you need to update test coverage, each package contains a `TestMain` function which handles test coverage, you should adjust this value only when strictly necessary. To run the unit tests use the `go test` command. `go test ./...` at the top level of the repo will run all unit tests for every folder. To run specific tests use the command `go test '-run=<REGEX_MATCHER>'` in the folder which the test resides. Appending `-coverprofile=coverage.out` to the test command will produce a file which can be viewed in a web browser using `go tool cover -html=coverage.out`.

This repo uses Godog to run cucumber functional tests, which are located in `internal/functionaltests`. The features for these tests are then located in the `features` folder. The functional tests use contracts from the `contracts` folder. To run the functional tests, run the `godog` command in the `internal/functionaltests` folder. You can find more information on Godog [here](https://github.com/cucumber/godog).

All source files in this repo require licenses at the top. You can find the text of this license [here](.golangci.yml). To perform license checking yourself, run `make lint`.

## Updating dependencies

If any changes are made to the module dependencies in [go.mod](go.mod), the module changes must be reflected in the integration test chaincodes, since they refer to the development codebase rather than a released version. This is done by running `make sync-deps`.

## Mechanics of Contributing
The codebase for this repo is maintained in GitHub, as such changes to the codebase should be given via a Pull Request.

The pull request title and commit messages should be a concise explanation of the changes being made. The PR should then contain more in depth information. If a change is requested you should add an additional commit and push this to the PR branch.

## Code of Conduct Guidelines
See our [Code of Conduct Guidelines](CODE_OF_CONDUCT.md)

## Maintainers
The maintainers of this repo can be found in the [Codeowners](CODEOWNERS.md) file.

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.