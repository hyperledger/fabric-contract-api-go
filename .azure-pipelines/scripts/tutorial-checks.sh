#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Go through each file in tutorial folder and test that all headers in file exist in contents
# and follow correct format

set -e

cd tutorials

check_contents_entries() {
    CONTENTS=$1

    NUM_CONTENT_ENTRIES=0
    while IFS= read -r CONTENTS_ENTRY; do
        if [[ ! "$CONTENTS_ENTRY" =~ ^[-\*][[:space:]]\[.*\]\(.*\) ]]; then
            echo "Invalid entry in $file contents: $CONTENTS_ENTRY"
            exit 1
        fi
        NUM_CONTENT_ENTRIES=$((NUM_CONTENT_ENTRIES + 1))
    done <<< "$CONTENTS"
}

check_num_entries() {
    if [ "$1" -ne "$2" ]; then
        echo "Additional entries in $file contents"
        exit 1
    fi
}

check_readme_contents_list() {
    file="$1"

    CONTENTS=$(cat $file | sed -n '/^## Contents/,/^##/p;/^## !Contents/q' | sed '1d' | sed '/^$/d')

    check_contents_entries "$CONTENTS"

    NUM_FILES=0
    for file in *; do
        if [ -f "$file" ]; then
            if [ "$file" != "README.md" ]; then
                FILE_FOUND=false
                while IFS= read -r CONTENTS_ENTRY; do
                    FORMATTED_FILE_NAME=$(echo "$file" | sed -e 's/-/ /g' | sed -e 's/\.md//g')
                    FORMATTED_FILE_NAME="$(tr '[:lower:]' '[:upper:]' <<< ${FORMATTED_FILE_NAME:0:1})${FORMATTED_FILE_NAME:1}"
                    if [[ "[$FORMATTED_FILE_NAME](./$file)" == "${CONTENTS_ENTRY/- /}" ]]; then
                        FILE_FOUND=true
                    fi
                done <<< "$CONTENTS"

                if [ "$FILE_FOUND" = false ]; then
                    echo "Missing file in README contents: $file"
                    exit 1
                fi

                NUM_FILES=$((NUM_FILES + 1))
            fi
        fi
    done

    check_num_entries $NUM_CONTENT_ENTRIES $NUM_FILES
}

check_contents_list() {
    file="$1"

    CONTENTS=$(cat $file | sed -n '/^## Tutorial contents/,/^##/p;/^## !Tutorial contents/q' | sed '1d;$d' | sed '/^$/d')
    HEADERS=$(cat $file | sed -n 's/## \(.*\)/\1/p' | sed '/^Tutorial contents/d' | sed '/^#/d')
    
    check_contents_entries "$CONTENTS"

    NUM_HEADERS=0
    while IFS= read -r HEADER; do
        HEADER_FOUND=false
        while IFS= read -r CONTENTS_ENTRY; do
            FORMATTED_HEADER=$(echo "${HEADER// /-}" | tr -cd '[:alnum:]-' | tr '[:upper:]' '[:lower:]')
            if [[ "[$HEADER](#$FORMATTED_HEADER)" == "${CONTENTS_ENTRY/- /}" ]]; then
                HEADER_FOUND=true
            fi
        done <<< "$CONTENTS"

        if [ "$HEADER_FOUND" = false ]; then
            echo "Missing header in $file contents: $HEADER"
            exit 1
        fi

        NUM_HEADERS=$((NUM_HEADERS + 1))
    done <<< "$HEADERS"

    check_num_entries $NUM_CONTENT_ENTRIES $NUM_HEADERS
}

for file in *; do
    if [ -f "$file" ]; then
       if [ "$file" != "README.md" ]; then
            check_contents_list "$file"
       else
            check_readme_contents_list "$file"
       fi
    fi
done
