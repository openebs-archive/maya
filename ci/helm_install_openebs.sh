#!/usr/bin/env bash

kubectl -n kube-system create sa tiller 
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller 
kubectl -n kube-system patch deploy/tiller-deploy -p '{"spec": {"template": {"spec": {"serviceAccountName": "tiller"}}}}' 
# With helm 2.9.0 and K8s 1.9.x there is an issue 
# Use the following workaround to enable access
# https://github.com/kubernetes/helm/issues/3985#issuecomment-385102874
kubectl -n kube-system patch deployment tiller-deploy -p '{"spec": {"template": {"spec": {"automountServiceAccountToken": true}}}}'

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
sleep 30
kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-storageclasses.yaml

