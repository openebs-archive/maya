#!bin/bash
# kubectl related generic functions
kubectlApply()
{
    # append all the files together prefixed with -f
    filesToApply=""
    while [ "$1" != "" ]; do
        filesToApply="$filesToApply -f $1"
        sleep 1
        shift
    done

    kubectl apply $filesToApply
}

kubectlDelete()
{
    # append all the files together prefixed with -f
    filesToDelete=""
    while [ "$1" != "" ]; do
        filesToDelete="$filesToDelete -f $1"
        sleep 1
        shift
    done

    kubectl delete $filesToDelete
}

# color related stuff

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'
