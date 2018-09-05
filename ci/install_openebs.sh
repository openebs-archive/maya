#!/usr/bin/env bash

echo "--------------------Installing openebs operator---------------------------"
sleep 5
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment -n openebs maya-apiserver -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for Maya-apiserver to be ready"
        sleep 10
    fi
done

echo "--------------- Maya apiserver initial logs -----------------------------"
MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
kubectl logs $MAPIPOD -n openebs
printf "\n\n"

#Print the default StoragePools Created
kubectl get sp

#Print the default StoragePoolClaim Created
kubectl get spc

#Print the default StorageClasses Created
kubectl get sc


for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment -n openebs openebs-provisioner -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for Openebs-provisioner to be ready"
        sleep 10
    fi
done

kubectl get pods --all-namespaces

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
kubectl logs --tail=200 $MAPIPOD -n openebs
printf "\n\n"

echo "--------------- Create Cstor and Jiva PersistentVolume ------------------"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/pvc-jiva-sc-1r.yaml
sleep 10
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/pvc-sparse-claim-cstor.yaml

sleep 30
echo "--------------------- List SC,PVC,PV and pods ---------------------------"
kubectl get sc,pvc,pv
kubectl get pods --all-namespaces

JIVACTRL=$(kubectl get deploy -l openebs.io/controller=jiva-controller --no-headers | awk {'print $1'})
for i in $(seq 1 5) ; do
    replicas=$(kubectl get deployment $JIVACTRL -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
      else
        echo "Waiting for volume ctrl to be ready"
        kubectl logs --tail=10 $MAPIPOD -n openebs
        printf "\n\n"
        sleep 30
    fi
done

JIVAREP=$(kubectl get deploy -l openebs.io/replica=jiva-replica --no-headers | awk {'print $1'})
for i in $(seq 1 5) ; do
    replicas=$(kubectl get deployment $JIVAREP -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
      else
        echo "Waiting for volume replica to be ready"
        kubectl logs --tail=10 $MAPIPOD -n openebs
        printf "\n\n"
        sleep 30
    fi
done

CSTORTARGET=$(kubectl get deploy -n openebs -l openebs.io/target=cstor-target --no-headers | awk {'print $1'})
for i in $(seq 1 5) ; do
    replicas=$(kubectl get deployment -n openebs $CSTORTARGET -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
      else
        echo "Waiting for cstor volume target to be ready"
        kubectl logs --tail=10 $MAPIPOD -n openebs
        printf "\n\n"
        sleep 30
    fi
done
