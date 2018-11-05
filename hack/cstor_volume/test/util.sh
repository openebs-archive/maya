#!bin/bash
# kubectl related generic functions

prefixedFileNames()
{
    # append all the files together prefixed with -f
    local filesToApply=""
    while [ "$1" != "" ]; do
        filesToApply="$filesToApply -f $1"
        sleep 1
        shift
    done

    # printing the following for N no of files passed as argument
    # -f file1 -f file2 -f file3 ... -f fileN
    echo $filesToApply
}

kubectlApply()
{
    filesToApply=$(prefixedFileNames $@)
    kubectl apply $filesToApply
}

kubectlDelete()
{
    filesToDelete=$(prefixedFileNames $@)
    kubectl delete $filesToDelete
}

# color related stuff

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'
