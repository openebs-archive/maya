#!/usr/bin/env bash

set -e
echo "" > coverage.txt
cd ..

for d in $(go list ./... | grep -v vendor); do
    #TODO - Include -race while creating the coverage profile. 
    go test -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
