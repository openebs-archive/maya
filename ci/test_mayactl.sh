#!/usr/bin/env bash

source ./ci/setup_env.sh

./ci/mayactl/run_vol_create.sh testvol
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

./ci/mayactl/run_vol_list.sh 
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

./ci/mayactl/run_vol_delete.sh testvol
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
