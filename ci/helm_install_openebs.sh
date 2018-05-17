#!/usr/bin/env bash

kubectl -n kube-system create sa tiller 
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller 
kubectl -n kube-system patch deploy/tiller-deploy -p '{"spec": {"template": {"spec": {"serviceAccountName": "tiller"}}}}' 

#Replace this with logic to wait till the pods are running
sleep 30
kubectl get pods --all-namespaces 
kubectl get sa

helm repo update
helm version
helm install stable/openebs --name ci --namespace openebs --set apiserver.imageTag="ci",jiva.replicas="1"

#Replace this with logic to wait till the pods are running
sleep 30
kubectl get pods --all-namespaces
