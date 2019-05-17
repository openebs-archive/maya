#!/usr/bin/env bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v 'vendor\|pkg/apis\|pkg/client/generated\|tests'); do
    #TODO - Include -race while creating the coverage profile.
    go test -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

echo "Running go vet on integration-test"
for d in $(go list ./... | grep 'tests'); do
    #TODO - Currently we are checking only on integration-test.
    go vet $d
done
