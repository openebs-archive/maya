#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs ci-openebs-maya-apiservice -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
