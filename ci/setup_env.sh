#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs maya-apiserver-service -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"

echo "*************** Running mayactl volume list *******************************"
${MAYACTL} volume list
#rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

POD=$(kubectl get pods -o=jsonpath='{.items[0].metadata.name}' -n openebs)
kubectl logs --tail=10 $POD -n openebs

printf "\n\n"

echo "************** Running mayactl volume info *******************************"
${MAYACTL} volume info --volname default-demo-vol1-claim
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

printf "\n\n"
sleep 10

echo "************** Running mayactl volume stats ******************************"
${MAYACTL} volume stats --volname default-demo-vol1-claim
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

#echo "************** Running mayactl volume delete ******************************"
#${MAYACTL} volume delete --volname demo-vol1-claim -n openebs
#rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

