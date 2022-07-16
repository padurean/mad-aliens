#!/bin/bash

set -e

# Skip the following packages/directories:
# /cmd/
echo "Get list of project pkgs, skipping /cmd/"
# NOTE: if more than one path will be added, they must be sepparated with \|
PKG_LIST=$(go list ./... | grep -v '/cmd/')
COVERAGE_DIR="${COVERAGE_DIR:-.coverage}"

echo "Run go mod tidy ..."
go mod tidy -v

echo "Run formating ..."
go fmt ./...

# Remove the directory containing the coverage output files
if [ -d "$COVERAGE_DIR" ]; then rm -Rf "$COVERAGE_DIR"; fi

# Test with -race flag by default

RACEFLAG="-race"
COVERMODE="atomic"

# Skip -race checks on CI
SKIPRACEFLAG=$1
if [ -n "$SKIPRACEFLAG" ]; then
    RACEFLAG=""
    COVERMODE="count"
fi
echo "RACEFLAG=$RACEFLAG"
echo "COVERMODE=$COVERMODE"

echo "Running tests and code coverage ..."

# Create the coverage files directory
mkdir -p "$COVERAGE_DIR";

# Required minimum coverage coverage
MINCOVERAGE=75

# Stop tests at first test fail
TFAILMARKER="FAIL:"
REGEXNOTFAILMARKER=".*no test files.*"
REGEXCOVERAGE="^coverage:"

for package in $PKG_LIST; do
    go test $RACEFLAG -covermode=$COVERMODE -coverprofile "${COVERAGE_DIR}/${package##*/}.cov" "$package" -v -count=1 -p=1 | { IFS=''; while read -r line; do
        echo "$line"

        if [ -z "$line" ]; then
            continue
        fi

        # To enforce that every package has tests, uncomment this block:
        # if [[ "${line}" =~ $REGEXNOTFAILMARKER ]] ; then
        #     echo ""
        #     echo "🚨 No tests for $package"
        #     echo ""
        #     exit 9
        # fi

        if [ -z "${line##*$TFAILMARKER*}" ] ; then
            exit 10
        fi

        if [[ "${line}" =~ $REGEXCOVERAGE ]] ; then
            pcoverage=$(echo "$line"| grep "coverage" | sed -E "s/.*coverage: ([0-9]*\.[0-9]+)\% of statements/\1/g")

            # To enforce a minimum coverage for package, uncomment this if:
            # if [ $(echo ${pcoverage%%.*}) -lt $MINCOVERAGE ] ; then
            #     echo ""
            #     echo "🚨 Test coverage of $package is $pcoverage%"
            #     echo "FAIL: min coverage is $MINCOVERAGE%"
            #     echo ""
            #     exit 11
            # else
                echo ""
                echo "🟢 Test coverage of $package is $pcoverage%"
                echo ""
            # fi
        fi
    done }
done

# Merge the coverage profile files
echo 'mode: count' > "${COVERAGE_DIR}"/coverage.cov
for fcov in "${COVERAGE_DIR}"/*.cov
do
    if [ $fcov != "${COVERAGE_DIR}/coverage.cov" ]; then
        tail -q -n +2 $fcov >> "${COVERAGE_DIR}"/coverage.cov
    fi
done


# Global code coverage
pcoverage=$(go tool cover -func="${COVERAGE_DIR}"/coverage.cov | grep 'total:' | sed -E "s/^total:.*\(statements\)[[:space:]]*([0-9]*\.[0-9]+)\%.*/\1/g")
echo "coverage: $pcoverage% of project"

# To enforce a minimum coverage for the project, uncomment this if:
# if [ $(echo ${pcoverage%%.*}) -lt $MINCOVERAGE ] ; then
#     echo ""
#     echo "🚨 Test coverage of project is $pcoverage%"
#     echo "FAIL: min coverage is $MINCOVERAGE%"
#     echo ""
#     exit 12
# else
    echo ""
    echo "🟢 Test coverage of project is $pcoverage%"
    echo ""
# fi
