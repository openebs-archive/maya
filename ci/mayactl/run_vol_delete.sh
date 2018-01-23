#!/usr/bin/env bash

${MAYACTL} volume delete --volname $1
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
