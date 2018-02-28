#!/usr/bin/env bash

kubectl -n kube-system create sa tiller 
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller 
kubectl -n kube-system patch deploy/tiller-deploy -p '{"spec": {"template": {"spec": {"serviceAccountName": "tiller"}}}}' 

#Replace this with logic to wait till the pods are running
sleep 30
kubectl get pods --all-namespaces 
kubectl get sa

helm repo add openebs-charts https://openebs.github.io/charts/
helm repo update
helm install openebs-charts/openebs --name ci --set apiserver.tag="ci",jiva.replicas="1"

#Replace this with logic to wait till the pods are running
sleep 30
kubectl get pods --all-namespaces
