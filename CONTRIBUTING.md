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

All source files in this repo require licenses at the top. You can find the text of this license [here](.azure-pipelines/resources/license.txt). License checking is performed using the npm package [license-check-and-add](https://www.npmjs.com/package/license-check-and-add) and config for this is located in `.azure-pipelines/resources/license-config.json`. To perform license checking yourself install the package globally and in the top level of this repo, run `license-check-and-add check -f .azure-pipelines/resources/license-config.json`.

## Mechanics of Contributing
The codebase for this repo is maintained in GitHub, as such changes to the codebase should be given via a Pull Request. An Azure Pipeline build is run against all pull requests and a passing build is required for code to be merged. The pipeline performs vetting, linting, license checking and testing. Issues for this repo are handled in [JIRA](https://jira.hyperledger.org). Fabric projects are split in JIRA and therefore all issues related to fabric-contract-api-go should use `FABCAG`. All pull requests should refer to a JIRA issue.

Pull requests should contain a single commit. The commit message should be prefixed with the issue number in square brackets e.g. `[FABCAG-XXXX]` followed by a concise explanation of the changes being made. The PR should then contain more in depth information. If a change is requested you should amend your original commit and force push over the original.

When you take on an issue raised in JIRA you should assign it to yourself and update that status as you do the work.

## Code of Conduct Guidelines
See our [Code of Conduct Guidelines](CODE_OF_CONDUCT.md)

## Maintainers
The maintainers of this repo can be found in the [Codeowners](CODEOWNERS.md) file.

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.