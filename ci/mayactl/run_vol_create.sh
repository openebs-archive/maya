#!/usr/bin/env bash

${MAYACTL} volume create --volname $1
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
