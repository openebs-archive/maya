#!bin/bash
source ./util.sh
echo -e ${GREEN}REMOVING OPENEBS${NC}

if [ "$POOL_FILE" == "" ]; then
    POOL_FILE="openebs-config.yaml"
fi

if [ "$OP_FILE" == "" ]; then
    OP_FILE="openebs-operator.yaml"
fi

kubectl delete po -n openebs -l "openebs.io/persistent-volume-claim=openebs-pvc" --force --grace-period=0

kubectlDelete $POOL_FILE
kubectlDelete $OP_FILE

kubectl delete crd castemplates.openebs.io cstorpools.openebs.io cstorvolumereplicas.openebs.io cstorvolumes.openebs.io disks.openebs.io runtasks.openebs.io storagepoolclaims.openebs.io  storagepools.openebs.io

