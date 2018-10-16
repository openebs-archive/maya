#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs maya-apiserver-service -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
export KUBECONFIG=$HOME/.kube/config

MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
CSTORVOL=$(kubectl get pv -o jsonpath='{.items[?(@.metadata.annotations.openebs\.io/cas-type=="cstor")].metadata.name}')
JIVAVOL=$(kubectl get pv -o jsonpath='{.items[?(@.metadata.annotations.openebs\.io/cas-type=="jiva")].metadata.name}')

echo "*************** Running mayactl volume list *******************************"
${MAYACTL} volume list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "************** Running Jiva mayactl volume describe **************************"
${MAYACTL} volume describe --volname $JIVAVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 5

echo "************** Running Jiva mayactl volume stats *************************"
${MAYACTL} volume stats --volname  $JIVAVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

#sleep 60
#echo "************** Running Jiva mayactl snapshot create **********************"
#${MAYACTL} snapshot create --volname $JIVAVOL --snapname snap1
#rc=$?;
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#printf "\n\n"
#sleep 30
#
#${MAYACTL} snapshot create --volname $JIVAVOL --snapname snap2
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#sleep 30
#
#echo "************** Running Jiva mayactl snapshot list ************************"
#${MAYACTL} snapshot list --volname $JIVAVOL
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#printf "\n\n"
#sleep 30
#echo "************** Running Jiva mayactl snapshot revert **********************"
#${MAYACTL} snapshot revert --volname $JIVAVOL --snapname snap1
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#printf "\n\n"
#sleep 10
#
#echo "************** Running Jiva mayactl snapshot list after revert ************"
#${MAYACTL} snapshot list --volname $JIVAVOL
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
echo "************** Running Jiva mayactl volume delete ************************"
${MAYACTL} volume delete --volname $JIVAVOL
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"
echo "************** Running Cstor mayactl volume describe *************************"
${MAYACTL} volume describe --volname $CSTORVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

