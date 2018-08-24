#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs maya-apiserver-service -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
export KUBECONFIG=$HOME/.kube/config
POD=$(kubectl get pods -o=jsonpath='{.items[0].metadata.name}' -n openebs)
PVNAME=$(kubectl get pv --no-headers | awk {'print $1'})

echo "*************** Running mayactl volume list *******************************"
${MAYACTL} volume list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "************** Running mayactl volume info *******************************"
${MAYACTL} volume info --volname $PVNAME
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 10

echo "************** Running mayactl volume stats ******************************"
${MAYACTL} volume stats --volname  $PVNAME
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

sleep 60
echo "************** Running mayactl snapshot create **************************"
${MAYACTL} snapshot create --volname $PVNAME --snapname snap1
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 30

${MAYACTL} snapshot create --volname $PVNAME --snapname snap2
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

sleep 30

echo "************** Running mayactl snapshot list ******************************"
${MAYACTL} snapshot list --volname $PVNAME
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 30
echo "************** Running mayactl snapshot revert ****************************"
${MAYACTL} snapshot revert --volname $PVNAME --snapname snap1
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi
printf "\n\n"
sleep 10

echo "************** Running mayactl snapshot list after revert ****************"
${MAYACTL} snapshot list --volname $PVNAME
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi
echo "************** Running mayactl volume delete ******************************"
${MAYACTL} volume delete --volname $PVNAME
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $POD -n openebs
	exit $rc;
fi

