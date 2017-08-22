#!/bin/bash
set -e

cd $(dirname $0)/../docs

#curl -s http://localhost:8080/v1/schemas?_role=project | jq . > ./input/schemas.json
curl -s http://localhost:8080/v2-beta/schemas?_role=project | jq . > ./input/schemas.json
echo Saved schemas.json

go run *.go -command=generate-collection-description
go run *.go -command=generate-description
go run *.go -command=generate-empty-description
go run *.go -command=generate-only-resource-fields

#go run *.go -command=generate-docs -version=v1.3 -lang=en -layout=rancher-api-v1-default-v1.3 -apiVersion=v1
go run *.go -command=generate-docs -version=v1.3 -lang=en -layout=rancher-api-v2-beta-default-v1.3 -apiVersion=v2-beta

echo Success
