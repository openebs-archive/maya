#!/usr/bin/env bash

MAPI_SVC_ADDR=`kubectl get service -n openebs maya-apiserver-service -o json | grep clusterIP | awk -F\" '{print $4}'`
export MAPI_ADDR="http://${MAPI_SVC_ADDR}:5656"
export KUBERNETES_SERVICE_HOST="127.0.0.1"
export MAYACTL="$GOPATH/src/github.com/openebs/maya/bin/maya/mayactl"
export KUBECONFIG=$HOME/.kube/config


MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
CSTORVOL=$(kubectl get pv -o jsonpath='{.items[?(@.metadata.annotations.openebs\.io/cas-type=="cstor")].metadata.name}')
JIVAVOL=$(kubectl get pv -o jsonpath='{.items[?(@.metadata.annotations.openebs\.io/cas-type=="jiva")].metadata.name}')
POOLNAME=$(kubectl get storagepools -o jsonpath='{.items[?(@.metadata.labels.openebs\.io/cas-type=="cstor")].metadata.name}')

function dumpMayaAPIServerLogs() {
  LC=$1
  MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
  kubectl logs --tail=${LC} $MAPIPOD -n openebs
  printf "\n\n"
}

echo "++++++++++++++++ Waiting for MAYA API's to get ready ++++++++++++++++++++++"
printf "\n\n"
echo "---------------- Checking Volume API \"/latest/volume\" -------------------"

for i in `seq 1 100`; do
    sleep 2
    responseCode=`curl -X GET --write-out %{http_code} --silent --output /dev/null $MAPI_ADDR/latest/volumes/`
    echo "Response Code from ApiServer: $responseCode"
    if [ $responseCode -ne 200 ]; then
        echo "Retrying.... $i"
        printf "Logs of api-server: \n\n"
        kubectl logs --tail=20 $MAPIPOD -n openebs
        printf "\n\n"
    else
        break
    fi
done

printf "\n\n"

echo "---------------- Checking Volume API \for jiva volume -------------------"
for i in `seq 1 100`; do
    sleep 2
    responseCode=`curl -X GET --write-out %{http_code} --silent --output /dev/null $MAPI_ADDR/latest/volumes/$JIVAVOL -H "namespace:default"`
    echo "Response Code from ApiServer: $responseCode"
    if [ $responseCode -ne 200 ]; then
        echo "Retrying.... $i"
        printf "Logs of api-server: \n\n"
        kubectl logs --tail=20 $MAPIPOD -n openebs
        printf "\n\n"
    else
        break
    fi
done

printf "\n\n"

echo "---------------- Checking Volume API \for cstor volume -------------------"
for i in `seq 1 100`; do
    sleep 2
    responseCode=`curl -X GET --write-out %{http_code} --silent --output /dev/null $MAPI_ADDR/latest/volumes/$CSTORVOL -H "namespace:openebs"`
    echo "Response Code from ApiServer: $responseCode"
    if [ $responseCode -ne 200 ]; then
        echo "Retrying.... $i"
        printf "Logs of api-server: \n\n"
        kubectl logs --tail=20 $MAPIPOD -n openebs
        printf "\n\n"
    else
        break
    fi
done

printf "\n\n"

echo "------------ Checking Volume STATS API \for cstor volume -----------------"
for i in `seq 1 100`; do
    sleep 2
    responseCode=`curl -X GET --write-out %{http_code} --silent --output /dev/null $MAPI_ADDR/latest/volumes/stats/$CSTORVOL -H "namespace:openebs"`
    echo "Response Code from ApiServer: $responseCode"
    if [ $responseCode -ne 200 ]; then
        echo "Retrying.... $i"
        printf "Logs of api-server: \n\n"
        kubectl logs --tail=20 $MAPIPOD -n openebs
        printf "\n\n"
    else
        break
    fi
done

printf "\n\n"

echo "------------ Checking Volume STATS API \for jiva volume -----------------"
for i in `seq 1 100`; do
    sleep 2
    responseCode=`curl -X GET --write-out %{http_code} --silent --output /dev/null $MAPI_ADDR/latest/volumes/stats/$JIVAVOL -H "namespace:default"`
    echo "Response Code from ApiServer: $responseCode"
    if [ $responseCode -ne 200 ]; then
        echo "Retrying.... $i"
        printf "Logs of api-server: \n\n"
        kubectl logs --tail=20 $MAPIPOD -n openebs
        printf "\n\n"
    else
        break
    fi
done

printf "\n\n"

echo "+++++++++++++++++++++ MAYA API's are ready ++++++++++++++++++++++++++++++++"

printf "\n\n"

echo "*************** Running mayactl pool list *******************************"
${MAYACTL} pool list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "*************** Running mayactl pool describe *******************************"
${MAYACTL} pool describe --poolname $POOLNAME
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"

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

echo "************** Running Cstor mayactl volume stats *************************"
${MAYACTL} volume stats --volname  $CSTORVOL -n openebs
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


printf "\n\n"
echo "************** Running Cstor mayactl volume describe *************************"
${MAYACTL} volume describe --volname $CSTORVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

echo "************** Snapshot and Clone related tests***************************"
# Create jiva volume for snapshot clone test ( cstor volume already exists)
#kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-single-replica-jiva.yaml

## install iscsi pkg
echo "Installing iscsi packages"
sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status

kubectl get pods --all-namespaces
kubectl get sc

sleep 30

echo "******************* Describe disks **************************"
kubectl describe disks

echo "******************* Describe spc,sp,csp **************************"
kubectl describe spc,sp,csp

echo "******************* List all pods **************************"
kubectl get po --all-namespaces

echo "******************* List PVC,PV and pods **************************"
kubectl get pvc,pv

# Create the application
echo "Creating busybox-jiva and busybox-cstor application pod"
kubectl create -f $GOPATH/src/github.com/openebs/maya/ci/snapshot/jiva/busybox.yaml
kubectl create -f $GOPATH/src/github.com/openebs/maya/ci/snapshot/cstor/busybox.yaml

for i in $(seq 1 100) ; do
    phaseJiva=$(kubectl get pods busybox-jiva --output="jsonpath={.status.phase}")
    phaseCstor=$(kubectl get pods busybox-cstor --output="jsonpath={.status.phase}")
    if [ "$phaseJiva" == "Running" ] && [ "$phaseCstor" == "Running" ]; then
        break
	else
        echo "busybox-jiva pod is in:" $phaseJiva
        echo "busybox-cstor pod is in:" $phaseCstor

        if [ "$phaseJiva" != "Running" ]; then
           kubectl describe pods busybox-jiva
        fi
        if [ "$phaseCstor" != "Running" ]; then
           kubectl describe pods busybox-cstor
        fi
        sleep 10
    fi
done

dumpMayaAPIServerLogs 100

echo "********************Creating volume snapshot*****************************"
kubectl create -f  $GOPATH/src/github.com/openebs/maya/ci/snapshot/jiva/snapshot.yaml
kubectl create -f  $GOPATH/src/github.com/openebs/maya/ci/snapshot/cstor/snapshot.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-controller

# It might take some time for cstor snapshot to get created. Wait for snapshot to get created
for i in $(seq 1 100) ; do
    kubectl get volumesnapshotdata
    count=$(kubectl get volumesnapshotdata | wc -l)
    # count should be 3 as one header line would also be present
    if [ "$count" == "3" ]; then
        break
    else
        echo "snapshot/(s) not created yet"
        kubectl get volumesnapshot,volumesnapshotdata
        sleep 10
    fi
done

kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-controller

# Promote/restore snapshot as persistent volume
sleep 30
echo "*****************Promoting snapshot as new PVC***************************"
kubectl create -f  $GOPATH/src/github.com/openebs/maya/ci/snapshot/jiva/snapshot_claim.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-provisioner
kubectl create -f  $GOPATH/src/github.com/openebs/maya/ci/snapshot/cstor/snapshot_claim.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-provisioner

sleep 30
# get clone replica pod IP to make a curl request to get the clone status
cloned_replica_ip=$(kubectl get pods -owide -l openebs.io/persistent-volume-claim=demo-snap-vol-claim-jiva --no-headers | grep -v ctrl | awk {'print $6'})
echo "***************** checking clone status *********************************"
for i in $(seq 1 5) ; do
		clonestatus=`curl http://$cloned_replica_ip:9502/v1/replicas/1 | jq '.clonestatus' | tr -d '"'`
		if [ "$clonestatus" == "completed" ]; then
            break
		else
            echo "Clone process in not completed ${clonestatus}"
            sleep 60
        fi
done

# Clone is in Alpha state, and kind of flaky sometimes, comment this integration test below for time being,
# util its stable in backend storage engine
echo "***************Creating busybox-clone-jiva application pod********************"
kubectl create -f $GOPATH/src/github.com/openebs/maya/ci/snapshot/jiva/busybox_clone.yaml
kubectl create -f $GOPATH/src/github.com/openebs/maya/ci/snapshot/cstor/busybox_clone.yaml


kubectl get pods --all-namespaces
kubectl get pvc --all-namespaces

for i in $(seq 1 15) ; do
    phaseJiva=$(kubectl get pods busybox-clone-jiva --output="jsonpath={.status.phase}")
    phaseCstor=$(kubectl get pods busybox-clone-cstor --output="jsonpath={.status.phase}")
    if [ "$phaseJiva" == "Running" ] && [ "$phaseCstor" == "Running" ]; then
        break
    else
        echo "busybox-clone-jiva pod is in:" $phaseJiva
        echo "busybox-clone-cstor pod is in:" $phaseCstor

        if [ "$phaseJiva" != "Running" ]; then
            kubectl describe pods busybox-clone-jiva
        fi
        if [ "$phaseCstor" != "Running" ]; then
            kubectl describe pods busybox-clone-cstor
        fi
		sleep 30
        fi
done


echo "********************** cvr status *************************"
kubectl get cvr -n openebs -o yaml

dumpMayaAPIServerLogs 100

kubectl get pods
kubectl get pvc

echo "*************Verifying data validity and Md5Sum Check********************"
hashjiva1=$(kubectl exec busybox-jiva -- md5sum /mnt/store1/date.txt | awk '{print $1}')
hashjiva2=$(kubectl exec busybox-clone-jiva -- md5sum /mnt/store2/date.txt | awk '{print $1}')

hashcstor1=$(kubectl exec busybox-cstor -- md5sum /mnt/store1/date.txt | awk '{print $1}')
hashcstor2=$(kubectl exec busybox-clone-cstor -- md5sum /mnt/store2/date.txt | awk '{print $1}')

echo "busybox jiva hash: $hashjiva1"
echo "busybox-clone-jiva hash: $hashjiva2"
echo "busybox cstor hash: $hashcstor1"
echo "busybox-clone-cstor hash: $hashcstor2"

if [ "$hashjiva1" != "" ] && [ "$hashcstor1" != "" ] && [ "$hashjiva1" == "$hashjiva2" ] && [ "$hashcstor1" == "$hashcstor2" ]; then
	echo "Md5Sum Check: PASSED"
else
    echo "Md5Sum Check: FAILED"; exit 1
fi


echo "************** Running Jiva mayactl volume delete ************************"
${MAYACTL} volume delete --volname $JIVAVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=100 $MAPIPOD -n openebs
	exit $rc;
fi
sleep 30

printf "\n\n"
echo "************** Check if jiva replica data is cleared *************************"
if [ -f /var/openebs/$JIVAVOL/volume.meta ]; then
	#Check if the job is in progress.
	printf "\n"
	ls -lR /var/openebs
	printf "\n"
	kubectl get jobs
	printf "\n"
	kubectl get pods
	printf "\n"
else
   echo "Jiva replica data is cleared successfully"
fi
