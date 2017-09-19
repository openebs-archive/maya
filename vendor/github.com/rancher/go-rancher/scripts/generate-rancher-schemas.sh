#!/bin/bash
set -e -x

usage()
{
   echo "Usage: $0 <schema_version> <url_to_schema>"
   echo "Usage: schema_version is v2 or v3. V2 is for rancher v1.6.x and V3 is for rancher v2.0"
   exit 1
}

cd $(dirname $0)/../generator

SCHEMA=$1
URL_BASE='http://localhost:8080'

if [ "$#" -gt 2 ]
then
  usage
fi

if [ "$1" != "v2" ] && [ "$1" != "v3" ]
then
  usage
fi

if [ "$2" != "" ]; then
    URL_BASE=$2
fi

if [ "$1" == "v2" ]
then
  SCHEMA='v2-beta'
fi

echo -n Waiting for cattle ${URL_BASE}/ping
while ! curl -fs ${URL_BASE}/ping; do
    echo -n .
    sleep 1
done
echo

source ../scripts/common_functions

gen ${URL_BASE}/v1-catalog catalog rename
gen "${URL_BASE}/$SCHEMA" $1

echo Success
