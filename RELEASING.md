# Release Checklist

Before releasing ensure that you have:
- Decided on a release tag. The repo uses a 3 number semantic versioning system and therefore the tag should be of the form v.X.X.X e.g. v1.0.0. This value will be known in the rest of this document as `<RELEASE_TAG>`.

Releasing:
- Go to: https://github.com/hyperledger/fabric-contract-api-go/releases
- Select "Draft a new release"
- Enter the tag as `<RELEASE_TAG>`.
- Give the release a title of "Release `<RELEASE_TAG>`"
- Add to the large textarea the release notes. These should consist of:
    - Include section called "Release Notes" containing the high-level view of changes made in the version that are of note.
    - (Optional) Include a section called "Migration Notes" detailing "gotchas" for the user of migrating to the new version.
    - (Optional) Include a section called "Bug Fixes" detailing important bug fixes in the version.
