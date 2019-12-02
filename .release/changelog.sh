#!/bin/sh

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

OLD_VERSION=""

if [ ! -z "$1" ]; then
    OLD_VERSION="${1}.."
fi

echo "## $2\n$(date)" >> CHANGELOG.new
echo "" >> CHANGELOG.new
git log ${OLD_VERSION}HEAD  --oneline | grep -v Merge | sed -e "s/\[\(FABCAG-[0-9]*\)\]/\[\1\](https:\/\/jira.hyperledger.org\/browse\/\1\)/" -e "s/ \(FABCAG-[0-9]*\)/ \[\1\](https:\/\/jira.hyperledger.org\/browse\/\1\)/" -e "s/\([0-9|a-z]*\)/* \[\1\](https:\/\/github.com\/hyperledger\/fabric-contract-api-go\/commit\/\1)/" >> CHANGELOG.new
echo "" >> CHANGELOG.new
cat CHANGELOG.md >> CHANGELOG.new
mv -f CHANGELOG.new CHANGELOG.md