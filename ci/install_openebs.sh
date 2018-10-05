#!/usr/bin/env bash

echo "--------------------Installing openebs operator---------------------------"
sleep 5

CI_BRANCH="master"
CI_TAG="ci"

#Images from this repo are always tagged as ci 
#The downloaded operator file will may contain a non-ci tag name 
# depending on when and from where it is being downloaded. For ex:
# - during the release time, the image tags can be versioned like 0.7.0-RC..
# - from a branch, the image tags can be the branch names like v0.7.x-ci
if [ ${CI_TAG} != "ci" ]; then
  sudo docker tag openebs/m-apiserver:ci openebs/m-apiserver:${CI_TAG}
  sudo docker tag openebs/m-exporter:ci openebs/m-exporter:${CI_TAG}
  sudo docker tag openebs/cstor-pool-mgmt:ci openebs/cstor-pool-mgmt:${CI_TAG}
  sudo docker tag openebs/cstor-volume-mgmt:ci openebs/cstor-volume-mgmt:${CI_TAG}
fi

kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/${CI_BRANCH}/k8s/openebs-operator.yaml

function waitForDeployment() {
  DEPLOY=$1
  NS=$2
  
  for i in $(seq 1 50) ; do
    kubectl get deployment -n ${NS} ${DEPLOY}
    replicas=$(kubectl get deployment -n ${NS} ${DEPLOY} -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
      break
    else
      echo "Waiting for ${DEPLOY} to be ready"
      if [ ${DEPLOY} != "maya-apiserver" ] && [ ${DEPLOY} != "openebs-provisioner" ]; then
        dumpMayaAPIServerLogs 10
      fi
      sleep 10
    fi
  done
}

function dumpMayaAPIServerLogs() {
  LC=$1
  MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
  kubectl logs --tail=${LC} $MAPIPOD -n openebs
  printf "\n\n"
}

waitForDeployment maya-apiserver openebs
waitForDeployment openebs-provisioner openebs
dumpMayaAPIServerLogs 200

kubectl get pods --all-namespaces


#Print the default StoragePools Created
kubectl get sp

#Print the default StoragePoolClaim Created
kubectl get spc

#Print the default StorageClasses Created
kubectl get sc



sleep 10
#echo "------------------ Deploy Pre-release features ---------------------------"
#kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-pre-release-features.yaml

echo "------------------------ Create sparse storagepoolclaim --------------- "
# delete the storagepoolclaim created earlier and create new spc with min/max pool
# count 1
kubectl delete spc --all
kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/spc-sparse-single.yaml
sleep 10

echo "--------------- Maya apiserver later logs -----------------------------"
dumpMayaAPIServerLogs 200

echo "--------------- Create Cstor and Jiva PersistentVolume ------------------"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/pvc-jiva-sc-1r.yaml
sleep 10
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/pvc-sparse-claim-cstor.yaml

sleep 30
echo "--------------------- List SC,PVC,PV and pods ---------------------------"
kubectl get sc,pvc,pv
kubectl get pods --all-namespaces

kubectl get deploy -l openebs.io/controller=jiva-controller
JIVACTRL=$(kubectl get deploy -l openebs.io/controller=jiva-controller --no-headers | awk {'print $1'})
waitForDeployment ${JIVACTRL} default 

kubectl get deploy -l openebs.io/replica=jiva-replica
JIVAREP=$(kubectl get deploy -l openebs.io/replica=jiva-replica --no-headers | awk {'print $1'})
waitForDeployment ${JIVAREP} default 

kubectl get deploy -n openebs -l openebs.io/target=cstor-target
CSTORTARGET=$(kubectl get deploy -n openebs -l openebs.io/target=cstor-target --no-headers | awk {'print $1'})
waitForDeployment ${CSTORTARGET} openebs
