#!/usr/bin/env bash

#OPENEBS="$(https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml)"
#CASTEMPLATE="${https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-cas-templates-pre-alpha.yaml}"
#OPENEBSVOLUME="${https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-standard-jiva-default.yaml}"

echo "--------------------Installing openebs operator---------------------------"
sleep 10
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

MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)

echo "--------------- Maya apiserver initial logs -----------------------------"
kubectl logs $MAPIPOD -n openebs
printf "\n\n"

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

#echo "------------Deploy CAS templates configuration for Maya-apiserver---------"
#kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-pre-release-features.yaml

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

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment default-demo-vol1-claim-ctrl -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for volume ctrl to be ready"
        sleep 10
    fi
done

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment default-demo-vol1-claim-rep -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting for volume replica to be ready"
        sleep 10
    fi
done

#echo "----------- Delete Persistentvolumeclaim and PersistentVolume ------------"
#kubectl delete -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-standard-jiva-default.yaml
