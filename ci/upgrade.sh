# !/usr/bin/env bash
set -e
Target_upgrade="0.8.1"
CI_BRANCH="v0.8.x"
UPGRADE_SOURCE="https://github.com/ashishranjan738/openebs/branches/upgrade/k8s/upgrades/0.8.0-0.8.1/"
UPGRADE_LOCATION="0.8.0-0.8.1"

# Installing openebs v0.8.x

printf "\n${Orange}Installing OpenEBS operator"
kubectl apply -f https://openebs.github.io/charts/openebs-operator-0.8.0.yaml

function printCurrentState() {
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
}

# Check whether all components are in Running state
while [[ $(kubectl get pods -n openebs --no-headers | awk '{print $3}' | grep -v Running | wc -l) -ne 0 ]]; do
    printCurrentState
done
# waitForDeployment waits for the deployment to get ready
function waitForDeployment() {
  DEPLOY=$1
  NS=$2

  for i in $(seq 1 50) ; do
    kubectl get deployment -n ${NS} ${DEPLOY}
    replicas=$(kubectl get deployment -n ${NS} ${DEPLOY} -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
      break
    else
      printf "\nWaiting for ${DEPLOY} to be ready\n\n"
      if [ ${DEPLOY} != "maya-apiserver" ] && [ ${DEPLOY} != "openebs-provisioner" ]; then
        dumpMayaAPIServerLogs 10
      fi
      printf "\n"
      sleep 10
    fi
  done
}

function dumpMayaAPIServerLogs() {
  printf "\nAPI Server logs\n\n"
  LC=$1
  MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
  kubectl logs --tail=${LC} $MAPIPOD -n openebs
  printf "\n\n"
}

waitForDeployment maya-apiserver openebs
waitForDeployment openebs-provisioner openebs
dumpMayaAPIServerLogs 200

printCurrentState

while kubectl get spc; [[ $? -ne 0 ]]; do
  printf "SPC resource not available.Retrying...\n"
  sleep 5
done

echo "------------------------ Create sparse storagepoolclaim --------------- "
# delete the storagepoolclaim created earlier and create new spc with min/max pool
# count 1
kubectl delete spc --all
while [[ `kubectl get deployments -l app=cstor-pool -n openebs --no-headers | wc -l` -ne 0 ]]; do 
  printf "Wating for old storage pool to get deleted\n"
  kubectl get deploy --all-namespaces
  sleep 5
done

kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/$CI_BRANCH/k8s/sample-pv-yamls/spc-sparse-single.yaml

while [[ `kubectl get deployments -l app=cstor-pool -n openebs --no-headers | wc -l` -ne 1 ]]; do 
  printf "Wating for new storage pool to get created\n"
  kubectl get deploy --all-namespaces
  sleep 5
done

pool_deployment=`kubectl get deploy  -n openebs -l app=cstor-pool -o jsonpath="{.items[*].metadata.name}"`

waitForDeployment $pool_deployment openebs

echo "--------------- Maya apiserver later logs -----------------------------"
dumpMayaAPIServerLogs 200

echo "--------------- Create Cstor and Jiva PersistentVolume ------------------"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/$CI_BRANCH/k8s/demo/pvc-single-replica-jiva.yaml
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/$CI_BRANCH/k8s/sample-pv-yamls/pvc-sparse-claim-cstor.yaml

printCurrentState


while [[ `kubectl get deploy -l openebs.io/controller=jiva-controller` == "" ]]; do
  kubectl get deploy -l openebs.io/controller=jiva-controller
  sleep 5
done

JIVACTRL=$(kubectl get deploy -l openebs.io/controller=jiva-controller --no-headers | awk {'print $1'})
waitForDeployment ${JIVACTRL} default

while [[ `kubectl get deploy -l openebs.io/replica=jiva-replica` == "" ]]; do
  kubectl get deploy -l openebs.io/replica=jiva-replica
  sleep 5
done

kubectl get deploy -l openebs.io/replica=jiva-replica
JIVAREP=$(kubectl get deploy -l openebs.io/replica=jiva-replica --no-headers | awk {'print $1'})
waitForDeployment ${JIVAREP} default

while [[ `kubectl get deploy -n openebs -l openebs.io/target=cstor-target` == "" ]]; do

  kubectl get deploy -n openebs -l openebs.io/target=cstor-target
  sleep 5
done

kubectl get deploy -n openebs -l openebs.io/target=cstor-target
CSTORTARGET=$(kubectl get deploy -n openebs -l openebs.io/target=cstor-target --no-headers | awk {'print $1'})
waitForDeployment ${CSTORTARGET} openebs


echo "---------------Creating in pvc---------------"

# Create the application
echo "Creating busybox-jiva and busybox-cstor application pod"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/ci/maya/snapshot/jiva/busybox.yaml
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/ci/maya/snapshot/cstor/busybox.yaml

kubectl get pods --all-namespaces
kubectl get pvc --all-namespaces

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
           printf "\n"
        fi
        if [ "$phaseCstor" != "Running" ]; then
           kubectl describe pods busybox-cstor
           printf "\n"
        fi
        sleep 10
    fi
done

JIVA_PVNAME=`kubectl get pv -l openebs.io/cas-type=jiva -o jsonpath="{.items[*].metadata.name}"`
CSTOR_PVNAME=`kubectl get pv -l openebs.io/cas-type=cstor -o jsonpath="{.items[*].metadata.name}"`
CSTOR_POOL=`kubectl get spc -o jsonpath="{.items[*].metadata.name}"`

# svn checkout $UPGRADE_SOURCE ci/
cd ci
chmod +x ./jiva_volume_update.sh
chmod +x ./cstor_pool_update.sh
chmod +x ./cstor_volume_update.sh

# Upgrade jiva volume
./jiva_volume_upgrade.sh $JIVA_PVNAME

# Upgrade cstor volumes
./cstor_pool_upgrade.sh $CSTOR_POOL openebs

# Upgrade cstor volume
./cstor_volume_upgrade.sh $CSTOR_PVNAME openebs

kubectl get deploy -l openebs.io/controller=jiva-controller
JIVACTRL=$(kubectl get deploy -l openebs.io/controller=jiva-controller --no-headers | awk {'print $1'})
waitForDeployment ${JIVACTRL} default

kubectl get deploy -l openebs.io/replica=jiva-replica
JIVAREP=$(kubectl get deploy -l openebs.io/replica=jiva-replica --no-headers | awk {'print $1'})
waitForDeployment ${JIVAREP} default

kubectl get deploy -n openebs -l openebs.io/target=cstor-target
CSTORTARGET=$(kubectl get deploy -n openebs -l openebs.io/target=cstor-target --no-headers | awk {'print $1'})
waitForDeployment ${CSTORTARGET} openebs


# Verify the cstor-pool-upgrade
if [ "`kubectl get csp -o jsonpath='{.items[*].metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in csp"
  exit 1
fi

if [ "`kubectl get sp -l openebs.io/cas-type=cstor -o jsonpath='{.items[*].metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in sp"
  exit 1
fi

if [ "`kubectl get deployments -n openebs -l app=cstor-pool -o jsonpath='{.items[*].metadata.labels.openebs\.io/version}'`" == ""  ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in deployment"
fi 


# Verify the jiva upgrade
jiva_controller_dep=$(echo $JIVA_PVNAME-ctrl)
jiva_replica_dep=$(echo $JIVA_PVNAME-rep)
jiva_controller_svc=$(echo $JIVA_PVNAME-ctrl-svc)


if [ "`kubectl get deploy $jiva_controller_dep -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref .metadata.annotation not found in jiva controller deployment $jiva_controller_dep"
  exit 1
fi

if [ "`kubectl get deploy $jiva_controller_dep -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in jiva controller deployment $jiva_controller_dep"
  exit 1
fi

if [ "`kubectl get deploy $jiva_controller_dep -o jsonpath='{.spec.template.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref .spec.template.metadata.annotation not found in jiva controller deployment $jiva_controller_dep"
  exit 1
fi

if [ "`kubectl get deploy $jiva_replica_dep -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref annotation not found in jiva replica deployment $jiva_replica_dep"
  exit 1
fi

if [ "`kubectl get deploy $jiva_replica_dep -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in jiva replica deployment $jiva_replica_dep"
  exit 1
fi

if [ "`kubectl get deploy $jiva_replica_dep -o jsonpath='{.spec.template.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref .spec.template.metadata.annotation not found in jiva replica deployment $jiva_replica_dep"
  exit 1
fi

if [ "`kubectl get svc $jiva_controller_svc -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref annotation not found in jiva service $jiva_controller_svc"
  exit 1
fi

if [ "`kubectl get svc $jiva_controller_svc -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade label not found in .metadata.labels in jiva controller svc $jiva_controller_svc"
  exit 1
fi




# Verify the cstor upgrade
cstor_target_dep=$(echo $CSTOR_PVNAME-target)
cstor_target_svc=$(echo $CSTOR_PVNAME)
cstor_target_vol=$(echo $CSTOR_PVNAME)
cstor_replicas=$(kubectl get cvr -n openebs -l cstorvolume.openebs.io/name=$CSTOR_PVNAME -o jsonpath="{range .items[*]}{@.metadata.name};{end}" | tr ";" "\n")

if [ "`kubectl get deploy $cstor_target_dep -n openebs -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref .metadata.annotation not found in cstor controller deployment $cstor_target_dep"
  exit 1
fi

if [ "`kubectl get deploy $cstor_target_dep -n openebs -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade .metadata.labels not found in cstor controller deployment $cstor_target_dep"
  exit 1
fi

if [ "`kubectl get deploy $cstor_target_dep -n openebs -o jsonpath='{.spec.template.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref .spec.template.metadata.annotation not found in cstor controller deployment $cstor_target_dep"
  exit 1
fi

if [ "`kubectl get svc $cstor_target_svc -n openebs -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref annotation not found in cstor service $cstor_target_svc"
  exit 1
fi

if [ "`kubectl get svc $cstor_target_svc -n openebs -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade label not found in .metadata.labels in cstor controller svc $cstor_target_svc"
  exit 1
fi

if [ "`kubectl get cstorvolume $cstor_target_vol -n openebs -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
  printf "openebs.io/storage-class-ref annotation not found in cstorvolume $cstor_target_vol"
  exit 1
fi

if [ "`kubectl get cstorvolume $cstor_target_vol -n openebs -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
  printf "openebs.io/version:$Target_upgrade label not found in .metadata.labels in cstorvolume $cstor_target_vol"
  exit 1
fi

for replica in $cstor_replicas
do
  if [ "`kubectl get cvr $replica -n openebs -o jsonpath='{.metadata.annotations.openebs\.io/storage-class-ref}'`" == "" ]; then
    printf "openebs.io/storage-class-ref annotation not found in cstorvolumereplica $replica"
    exit 1
  fi

  if [ "`kubectl get cvr $replica -n openebs -o jsonpath='{.metadata.labels.openebs\.io/version}'`" != "$Target_upgrade" ]; then
    printf "openebs.io/version:$Target_upgrade label not found in .metadata.labels in cstorvolumereplica $replica"
    exit 1
  fi
done



