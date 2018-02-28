#!/usr/bin/env bash
MAPI_SERVICE=`kubectl get svc -o jsonpath="{range.items[?(@.spec.ports[*].port==5656)]}{.metadata.name}:{end}"`
MAPI_SVC_ADDR=`kubectl get service $MAPI_SERVICE -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
