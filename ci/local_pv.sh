#!/usr/bin/env bash

ARTIFACTS_DIR="$GOPATH/src/github.com/openebs/maya/integration-tests/artifacts"
kubectl apply -f ${ARTIFACTS_DIR}/openebs-local-provisioner.yaml
kubectl apply -f ${ARTIFACTS_DIR}/sc-hostpath.yaml
sleep 30
LOCALPV="INIT"
for _ in $(seq 1 5) ; do
    phaseLocalPV=$(kubectl get pods busybox-hp --output="jsonpath={.status.phase}")
    if [ "$phaseLocalPV" == "Running" ]; then
        LOCALPV="RUNNING"
        break
    else
        echo "busybox-hp is in:" $phaseLocalPV
        if [ "$phaseLocalPV" != "Running" ]; then
            kubectl describe pods busybox-hp
        fi
	sleep 30
    fi
done

if [ "$LOCALPV" != "RUNNING" ]; then
	echo "Failed Local PV Tests"
	exit 1
else
	echo "Completed Local PV Tests"
fi
