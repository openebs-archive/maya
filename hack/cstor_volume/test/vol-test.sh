#!bin/bash

# this script sets up openebs and then calls different test scripts
# source util file
source util.sh

usage()
{
    echo bash vol-test.sh -o operator.yaml -o pool.yaml
}

runTest()
{
    for testCase in test_*; do
        echo "-----------------------------------------------------"
        echo -e ${YELLOW}Running $testCase${NC}
        echo "-----------------------------------------------------"
        cd $testCase
        echo -e ${YELLOW}`cat description`${NC}
        bash test-script.sh
        exitValue=$?
        echo "-----------------------------------------------------"
        if [ "$exitValue" == "0" ]; then
            echo -e ${GREEN}Success: $testCase${NC}
        else
            echo -e ${RED}Fail: $testCase${NC}
        fi
        sleep 10
        echo "Deleting residual pods forcefully if any"
        kubectl delete po -n openebs -l "openebs.io/persistent-volume-claim=openebs-pvc" --force --grace-period=0
        cd ..
    done
}
# Setup operator
setupOperator()
{
    opYaml="$1"
    echo using operator file $opYaml
    kubectl apply -f $opYaml
}

# Setup pool
setupPool()
{
    poolYaml="$1"
    echo using pool file $poolYaml
    kubectl apply -f $poolYaml
}

# get operator yaml and pool yaml
while getopts "o:p:" opt; do
    case "${opt}" in
        o)
            OP_FILE="${OPTARG}"
            ;;
        p)
            POOL_FILE="${OPTARG}"
            ;;
        :)
            echo option -$OPTARG requires argument
            exit 1
            ;;
    esac
done

if [ "$OP_FILE" == "" ] || [ "$POOL_FILE" == "" ]; then
    usage
    exit 1
fi

# exporting so that this could be used during cleanup
export OP_FILE=$OP_FILE
export POOL_FILE=$POOL_FILE

echo Received operator $OP_FILE and pool $POOL_FILE

setupOperator $OP_FILE

# check if the maya api server is in running state yet. If m-api server is not in running state then
# retry. This is being done to ensure that the installer installs all the resources before we proceed
# further
sleep 10
try=1
appStatus=
printf "%s" "Checking status of maya api server:"
until [ "$appStatus" == "Running"  ] || [ $try == 30 ]; do
    printf " %s" "$try"
    appStatus=$(kubectl get po -l name=maya-apiserver -n openebs -o jsonpath='{.items[0].status.phase}')
    try=`expr $try + 1`
    sleep 5
done

echo ""

if [ "$appStatus" != "Running" ]; then
    echo Maya api server pod not up yet. Exiting...
    sleep 10
    exit 0
fi

echo Maya-apiserver is running

setupPool $POOL_FILE

# Call test cases
# go inside each test case directory. execute the test script and then cd back
runTest

bash ./delete.sh
