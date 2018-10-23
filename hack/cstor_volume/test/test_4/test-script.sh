#!bin/bash

source ../util.sh

reincarnate()
{
    echo Scaling down cstor pool deploys replica to 0
    kubectl scale deploy -l openebs.io/storage-pool-claim=cstor-pool-default-0.7.0 --replicas=0 -n openebs
    if [ "$?" != "0" ]; then
        exit $?
    fi
    sleep 15

    try=1
    aliveTargets=
    printf "%s " "Checking if target pods are killed"
    until [ "$aliveTargets" == "0" ] || [ $try == 10 ]; do
        printf "%s " $try
        aliveTargets=`kubectl get po -l openebs.io/storage-pool-claim=cstor-pool-default-0.7.0 -n openebs -o jsonpath='{.items[?(@.status.phase=="Running")].status.phase}' | wc -l`
        try=`expr $try + 1`
        sleep 5
    done
    echo ""

    if [ $try == 10 ]; then
        echo Target could not be killed in given duration
        cleanUp
        exit 1
    fi
    echo "All target pods killed"
    echo Scaling up cstor pool deploys count back to 1
    kubectl scale deploy -l openebs.io/storage-pool-claim=cstor-pool-default-0.7.0 --replicas=1 -n openebs
    if [ "$?" != "0" ]; then
        exit $?
    fi
    sleep 10
}

cleanUp()
{
    kubectlDelete app.yaml
    kubectlDelete pvc.yaml
    sleep 5
    kubectlDelete sc.yaml
}
# Apply storage class
echo Applying sc
kubectlApply sc.yaml

# Apply the pvc.yaml and then application.yaml

echo Applying pvc
kubectlApply pvc.yaml

echo Applying app
kubectlApply app.yaml

# sleep as image pulling takes time
sleep 10
appStatus=
try=1
printf "%s:" "Checking status of application"
until [ "$appStatus" == "Running"  ] || [ $try == 30 ]; do
    printf "%s " $try
    appStatus=$(kubectl get po -l name=nginx -o jsonpath='{.items[0].status.phase}')
    try=`expr $try + 1`
    sleep 5
done
echo ""
if [ "$appStatus" == "Running" ]; then
    echo Application is in running state
    reincarnate
else
    echo Application did not come up.
    cleanUp
    exit 1
fi


try=1
writeStatus=

export appName=`kubectl get po -l name=nginx -o jsonpath='{.items[0].metadata.name}'`
printf "%s" "Trying to write file to openebs vol in the app"
until [ "$writeStatus" == "0"  ] || [ $try == 20 ]; do
    printf " %s" $try
    # run a separate process and kill it after some time
    # the script would write its exit status into a file
    # write-status.txt. 0 value indicates success any other
    # value is failure
    bash sanity-script.sh &
    pid=$!
    sleep 8
    kill $pid > /dev/null 2>&1
    writeStatus=`cat write-status.txt`
    try=`expr $try + 1`
done

echo ""

if [ "$writeStatus" == "0" ]; then
    echo File creation was successful in openebs vol
    cleanUp
    exit 0
fi

echo File creation failed in openebs vol
cleanUp
exit 1
