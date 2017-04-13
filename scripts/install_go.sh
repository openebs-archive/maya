#!/bin/bash

# Used by development setups especially Vagrant
# NOTE: Use of vagrant user !!!
set -ex

GO_VERSION="1.8"
CURDIR=`pwd`

# Setup go, for development
SRCROOT="/opt/go"
SRCPATH="/opt/gopath"

# Get the ARCH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

# updating GO_VERSION if new version is available
#func to compare versions
version_gt()
{ 
	test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1";
}

CONTENT=$(wget https://storage.googleapis.com/golang -q -O -)
GO_LATEST=`echo -n $CONTENT | grep -o "go[0-9]\.[0-9][\.]*[0-9]*\.linux-${ARCH}.tar.gz" | grep -o "[0-9]\.[0-9][\.]*[0-9]*\." | sort --version-sort | tail -1 | head -c -2`

# updating GO_VERSION
if version_gt $GO_LATEST $GO_VERSION; then
     GO_VERSION=$GO_LATEST
fi

# Install Go
cd /tmp

if [ ! -f "./go${GO_VERSION}.linux-${ARCH}.tar.gz" ]; then
  wget -q https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${ARCH}.tar.gz
fi

tar -xf go${GO_VERSION}.linux-${ARCH}.tar.gz
sudo mv go $SRCROOT
sudo chmod 775 $SRCROOT
sudo chown vagrant:vagrant $SRCROOT

# Setup the GOPATH; 
# This allows subsequent "go get" commands to work.
sudo mkdir -p $SRCPATH
sudo chown -R vagrant:vagrant $SRCPATH 2>/dev/null || true
# ^^ silencing errors here because we expect this to fail for the shared folder

cat <<EOF >/tmp/gopath.sh
export GOPATH="$SRCPATH"
export GOROOT="$SRCROOT"
export PATH="$SRCROOT/bin:$SRCPATH/bin:\$PATH"
EOF

sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh

cd ${CURDIR}
