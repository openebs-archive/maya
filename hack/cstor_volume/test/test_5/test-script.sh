#!bin/bash

source ../util.sh
# TODO: Copy csp status test from ashutosh's code

# cleans up all the yamls that were applied intitially
cleanUp()
{
    kubectlDelete app.yaml
    kubectlDelete pvc.yaml
    # sleep as pvc deletion is dependent on sc
    sleep 10
    kubectlDelete sc.yaml
    sleep 5
}

# check if all the dependent resources are cleaned up. the count of each resource should be 0
checkDependentResources()
{
    pvName=`kubectl get pvc openebs-pvc -o jsonpath='{.spec.volumeName}'`
    pvSelector="-l openebs.io/persistent-volume=$pvName"
    try=1
    # initialize all count to empty initially
    csvCount=
    cvrCount=
    svcCount=
    targetDeployCount=
    targetPoCount=
    printf "%s" "Checking cStor volume artifacts:"
    until ( [ "$csvCount" == "0" ] && [ "$cvrCount" == "0" ] && [ "$svcCount" == "0" ] && [ "$targetDeployCount" == "0" ] ) || [ "$try" == "25" ]; do
        printf " %s" "$try"
        csvCount=`kubectl get cstorvolume $pvSelector -n openebs -o jsonpath='{.items[*].metadata.name}' | grep "" | wc -l`
        cvrCount=`kubectl get cvr -n openebs $pvSelector -o jsonpath='{.items[*].metadata.name}' | grep "" | wc -l`
        svcCount=`kubectl get svc -l openebs.io/target-service=cstor-target-svc $pvSelector -n openebs -o jsonpath='{.items[*].metadata.name}' | grep "" | wc -l`
        targetDeployCount=`kubectl get deploy -n openebs $pvSelector -l openebs.io/target=cstor-target -o jsonpath='{.items[*].metadata.name}' | grep "" | wc -l`
        targetPoCount=`kubectl get po -n openebs $pvSelector -l openebs.io/target=cstor-target -o jsonpath='{.items[*].metadata.name}' | grep "" | wc -l`

        try=`expr $try + 1`
        sleep 5
    done
    echo ""
    echo Resources status- csvCount: $csvCount, cvrCount: $cvrCount, svcCount: $svcCount, targetDeployCount: $targetDeployCount, targetPoCount: $targetPoCount \(Ignored\)
    if [ "$try" == "25" ]; then
        echo Resources are not cleaned up
        exit 1
    else
        echo All dependent artifacts deleted
    fi
}
# Apply storage class
echo Applying sc
kubectlApply sc.yaml

kubectlApply service-account.yaml

# Apply the pvc.yaml and then application.yaml

echo Applying pvc
kubectlApply pvc.yaml

echo Applying app
kubectlApply app.yaml

# sleep as image pulling takes time
sleep 10
appStatus=
try=1
printf "%s" "Checking status of application"
until [ "$appStatus" == "Running"  ] || [ $try == 30 ]; do
    printf " %s" $try
    appStatus=$(kubectl get po -l name=nginx -o jsonpath='{.items[0].status.phase}')
    try=`expr $try + 1`
    sleep 6
done
echo ""
workloadNode=$(kubectl get po -l name=nginx -o jsonpath='{.items[0].nodeName}')
targetNode=$(kubectl get po -l openebs.io/persistent-volume-claim=openebs-pvc -o jsonpath='{.items[0].nodeName}' -n openebs)

echo Workload is scheduled on $workloadNode
echo Target pod is scheduled on $targetNode

if [ "$appStatus" == "Running" ] && [ "$workloadNode" == "$targetNode" ]; then
    echo Application is in Running state
    echo Deleting the cStor Volume
    cleanUp
    checkDependentResources
    exit 0
else
    cleanUp
fi

exit 1
