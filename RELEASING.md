# Release Checklist

Before releasing ensure that you have:
- Decided on a release tag. The repo uses a 3 number semantic versioning system and therefore the tag should be of the form v.X.X.X e.g. v1.0.0. This value will be known in the rest of this document as `<RELEASE_TAG>`.
- Run `.release/changelog.sh <PREVIOUS_RELEASE_TAG> <RELEASE_TAG>` to update CHANGELOG.md to contain all commits since the previous tag. Running without the previous release tag will get all commits.
- Commit the updated changelog via PR with commit message "Preparing for release `<RELEASE_TAG>`"

Releasing:
- Go to: https://github.com/hyperledger/fabric-contract-api-go/releases
- Select "Draft a new release"
- Enter the tag as `<RELEASE_TAG>`.
- Give the release a title of "Release `<RELEASE_TAG>`"
- Add to the large textarea the release notes. These should consist of:
    - Include section called "Release Notes" containing the high-level view of changes made in the version that are of note.
    - (Optional) Include a section called "Migration Notes" detailing "gotchas" for the user of migrating to the new version.
    - (Optional) Include a section called "Bug Fixes" detailing important bug fixes in the version. Should be listed as bullet points with a link to the JIRA for that bug.
