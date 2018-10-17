#!/usr/bin/env bash
#
# This script runs tests and generates a report file.

set -e
echo "" > coverage.txt
PACKAGES=$(go list ./... | grep -Ev 'example|/vendor/')
for d in $PACKAGES; do
	go test -coverprofile=profile.out -covermode=atomic $d
	if [ -f profile.out ]; then
		cat profile.out >> coverage.txt
		rm profile.out
	fi
done

