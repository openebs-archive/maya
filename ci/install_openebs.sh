#!/usr/bin/env bash

#OPENEBS="$(https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml)"
#CASTEMPLATE="${https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-cas-templates-pre-alpha.yaml}"
#OPENEBSVOLUME="${https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-standard-jiva-default.yaml}"

echo "--------------------Installing openebs operator---------------------------"
sleep 10
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml

#cStor tests are not yet enabled on travis. 
#Deleting the NDM that gets installed by default. 
kubectl delete ds -n openebs node-disk-manager

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

#echo "----------Deploy Pre-release features---------"
#kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-pre-release-features.yaml

echo "--------------- Maya apiserver later logs -----------------------------"
kubectl logs --tail=200 $MAPIPOD -n openebs
printf "\n\n"

sleep 10
echo "-----------Create Persistentvolumeclaim and PersistentVolume ------------"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-single-replica-jiva.yaml

sleep 30
echo "--------------------- List PVC,PV and pods ---------------------------"
kubectl get pvc,pv

kubectl get pods --all-namespaces

JIVACTRL=$(kubectl get deploy -l openebs.io/controller=jiva-controller --no-headers | awk {'print $1'})
for i in $(seq 1 5) ; do
    replicas=$(kubectl get deployment $JIVACTRL -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for volume ctrl to be ready"
        kubectl logs $MAPIPOD -n openebs
        printf "\n\n"
        sleep 60
    fi
done

JIVAREP=$(kubectl get deploy -l openebs.io/replica=jiva-replica --no-headers | awk {'print $1'})
for i in $(seq 1 5) ; do
    replicas=$(kubectl get deployment $JIVAREP -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for volume replica to be ready"
        kubectl logs $MAPIPOD -n openebs
        printf "\n\n"
        sleep 60
    fi
done

#echo "----------- Delete Persistentvolumeclaim and PersistentVolume ------------"
#kubectl delete -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-standard-jiva-default.yaml
