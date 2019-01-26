#!/usr/bin/env bash
set -e
CYAN='\033[0;36m'
NC='\033[0m'
RED='\033[0;31m' 
ORANGE='\033[0;33m'

# Check for openebs namespace
printf "\n${ORANGE}Check for openebs namespace ${NC}\n";

if [[ $(kubectl get ns | grep openebs | awk '{ print $1}') != "openebs" ]]; then
    printf "\n${RED}openebs namespace not found ${NC}\n";
    exit 1
fi

# Current Cluster state
printf "\n*********Current state of cluster*********\n";

printf "\n${CYAN}Output of 'kubectl get pods --all-namespaces':${NC} \n\n"

kubectl get pods --all-namespaces

printf "\n${CYAN}Output of 'kubectl get svc --all-namespaces':${NC}\n\n"
kubectl get svc --all-namespaces
 
printf "\n${CYAN}Output of 'kubectl get deployment --all-namespaces':${NC}\n\n"
kubectl get deployment --all-namespaces

printf "\n${CYAN}Output of 'kubectl get pvc --all-namespaces':${NC}\n\n"
kubectl get pvc --all-namespaces

printf "\n${CYAN}Output of 'kubectl get pv --all-namespaces':${NC}\n\n"
kubectl get pv

printf "\n*******************************************\n";

kubectl delete pvc --all

# Removing all openebs components
printf "\n${ORANGE}Cleaning up OpenEBS components${NC}\n\n"
kubectl delete -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml


while [[ $(kubectl get ns | grep openebs | awk '{ print $1}') == "openebs" ]]; do 
    # Current Cluster state
    printf "\n*********Current state of cluster*********\n";

    printf "\n${CYAN}Output of 'kubectl get pods --all-namespaces':${NC} \n\n"

    kubectl get pods --all-namespaces

    printf "\n${CYAN}Output of 'kubectl get svc --all-namespaces':${NC}\n\n"
    kubectl get svc --all-namespaces
    
    printf "\n${CYAN}Output of 'kubectl get deployment --all-namespaces':${NC}\n\n"
    kubectl get deployment --all-namespaces

    printf "\n${CYAN}Output of 'kubectl get pvc --all-namespaces':${NC}\n\n"
    kubectl get pvc --all-namespaces

    printf "\n${CYAN}Output of 'kubectl get pv --all-namespaces':${NC}\n\n"
    kubectl get pv

    printf "\n${CYAN}Output of 'kubectl get ns':${NC}\n\n"
    kubectl get pv

    printf "\n*******************************************\n";
    sleep 5
done