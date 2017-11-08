#!/usr/bin/env bash

${MAYACTL} volume list
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
