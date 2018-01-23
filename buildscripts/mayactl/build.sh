#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../../" && pwd )"

# Change into that directory
cd "$DIR"

# Get the git commit
if [ -f $GOPATH/src/github.com/openebs/maya/GITCOMMIT ];
then
	GIT_CIMMIT="`cat $GOPATH/src/github.com/openebs/maya/GITCOMMIT`"
else
	GIT_COMMIT="$(git rev-parse HEAD)"
fi

# Get the version details
VERSION="`cat $GOPATH/src/github.com/openebs/maya/VERSION`"
VERSION_META="`cat $GOPATH/src/github.com/openebs/maya/BUILDMETA`"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64"}
XC_OS=${XC_OS:-"linux"}

XC_ARCHS=(${XC_ARCH// / })
XC_OSS=(${XC_OS// / })

# Delete the old dir
echo "==> Removing old directory..."
rm -rf bin/maya/*
mkdir -p bin/maya/

if [ -z "${MAYACTL}" ];
then
    MAYACTL="mayactl"
fi

# If its dev mode, only build for ourself
if [[ "${MAYA_DEV}" ]]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

# Build!
echo "==> Building..."
for GOOS in "${XC_OSS[@]}"
do
    for GOARCH in "${XC_ARCHS[@]}"
    do
        output_name="bin/maya/"$GOOS"_"$GOARCH"/"$MAYACTL

        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi
        env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags \
           "-X github.com/openebs/maya/pkg/version.GitCommit=${GIT_COMMIT} \
            -X main.CtlName='${CTLNAME}' \
            -X github.com/openebs/maya/pkg/version.Version=${VERSION} \
            -X github.com/openebs/maya/pkg/version.VersionMeta=${VERSION_META}"\
            -o $output_name\
            ./cmd/mayactl
    done

done

echo ""

# Move all the compiled things to the $GOPATH/bin
GOPATH=${GOPATH:-$(go env GOPATH)}
case $(uname) in
    CYGWIN*)
        GOPATH="$(cygpath $GOPATH)"
        ;;
esac
OLDIFS=$IFS
IFS=: MAIN_GOPATH=($GOPATH)
IFS=$OLDIFS

# Create the gopath bin if not already available
mkdir -p ${MAIN_GOPATH}/bin/

# Copy our OS/Arch to the bin/ directory
DEV_PLATFORM="./bin/maya/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
    cp ${F} bin/maya/
    cp ${F} ${MAIN_GOPATH}/bin/
done

if [[ "x${MAYA_DEV}" == "x" ]]; then
    # Zip and copy to the dist dir
    echo "==> Packaging..."
    for PLATFORM in $(find ./bin/maya -mindepth 1 -maxdepth 1 -type d); do
        OSARCH=$(basename ${PLATFORM})
        echo "--> ${OSARCH}"

        pushd $PLATFORM >/dev/null 2>&1
        zip ../maya-${OSARCH}.zip ./*
        popd >/dev/null 2>&1
    done
fi

# Done!
echo
echo "==> Results:"
ls -hl bin/maya/
