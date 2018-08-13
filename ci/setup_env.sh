#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs maya-apiserver-service -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
export KUBECONFIG=$HOME/.kube/config
POD=$(kubectl get pods -o=jsonpath='{.items[0].metadata.name}' -n openebs)
SNAPNAME=$(printf "%s_%s" "testsnap" "$(date +%F%N)")

echo "*************** Running mayactl volume list *******************************"
${MAYACTL} volume list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "************** Running mayactl volume info *******************************"
${MAYACTL} volume info --volname default-demo-vol1-claim
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 10

echo "************** Running mayactl volume stats ******************************"
${MAYACTL} volume stats --volname default-demo-vol1-claim
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

echo "************** Running mayactl snapshot create **************************"
${MAYACTL} snapshot create --volname default-demo-vol1-claim --snapname $SNAPNAME
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 30

${MAYACTL} snapshot create --volname default-demo-vol1-claim --snapname snap2
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

sleep 30

echo "************** Running mayactl snapshot list ******************************"
${MAYACTL} snapshot list --volname default-demo-vol1-claim
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 5
echo "************** Running mayactl volume delete ******************************"
${MAYACTL} volume delete --volname default-demo-vol1-claim
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

