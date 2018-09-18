namespace=openebs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
run_test(){
clusterEnvTemp=`cat $2 | awk 'NR==4 {print}'`
clusterEnv=${clusterEnvTemp##*:}
if [ "$clusterEnv" == "true" ];then
# Run commands to generate proper cluster environment
getCommandTemp=`cat $2 | awk 'NR==5 {print}'`
getCommand=${getCommandTemp##*:}
echo $getCommand
eval $getCommand
fi

resetClusterEnvTemp=`cat $2 | awk 'NR==7 {print}'`
resetClusterEnv=${resetClusterEnvTemp##*:}
if [ "$resetClusterEnv" == "true" ];then
getCommandTemp=`cat $2 | awk 'NR==8 {print}'`
echo $getCommandTemp
getCommand=${getCommandTemp##*:}
echo $getCommand
eval $getCommand
fi

applySpcTemp=`cat $2 | awk 'NR==9 {print}'`
applySpc=${applySpcTemp##*:}
    if [ "$applySpc" != "false"  ];then
        echo "Applying SPC Yaml"
        temp1=$(kubectl apply -f $1)
        echo "$temp1"
        temp2=${temp1##*/}
        spcName=${temp2%" "*}
        echo "SPC name is" $spcName

    fi

commandAfterSpcTemp=`cat $2 | awk 'NR==10 {print}'`
commandAfterSpc=${commandAfterSpcTemp##*:}
if [ "$commandAfterSpc" == "true" ];then
getCommandTemp=`cat $2 | awk 'NR==11 {print}'`
echo $getCommandTemp
getCommand=${getCommandTemp##*:}
echo $getCommand
eval $getCommand
fi

testCaseStatusFlag=0
testCaseTypeTemp=`cat $2 | awk 'NR==1 {print}'`
testCaseType=${testCaseTypeTemp##*:}
if [ "$testCaseType" == "negative" ];then
    check_csp_count $spcName $2
else
check_csp_count $spcName $2
check_deploy_count $spcName $2
check_sp_count $spcName $2
check_pool_status $spcName $2
fi
return $testCaseStatusFlag
}
check_pool_status(){
maxRetry=10
expectedPoolStatusTemp=`cat $2 | awk 'NR==3 {print}'`
expectedPoolStatus=${expectedPoolStatusTemp##*:}
local poolstatus=-1
until [ "$poolstatus" == "$expectedPoolStatus" ]
do
    if [ $maxRetry == 0 ];then
        break
    fi
    poolstatus=$(kubectl get csp -l openebs.io/storage-pool-claim=$1 -o=jsonpath='{range .items[*]}{@.status.phase}{end}'| grep -o "Online"| wc -l)
    echo -n "->"
    maxRetry=`expr $maxRetry - 1`
    sleep 10;
done

if [ "$poolstatus" == "$expectedPoolStatus" ]; then
    echo "Online CSP count is equal to expected count"
    testCaseStatusFlag=1
else
    echo "Online CSP count is not equal to expected count"
    testCaseStatusFlag=0
fi
}
check_sp_count(){
expectedSpCountTemp=`cat $2 | awk 'NR==2 {print}'`
expectedSpCount=${expectedSpCountTemp##*:}
maxRetry=10
local spCount=-1
until [ "$spCount" == "$expectedSpCount" ]
do
    if [ $maxRetry == 0 ];then
        break
    fi
    spCount=$(kubectl get sp --no-headers -l openebs.io/storage-pool-claim=$1|wc -l)
    echo -n "->"
    maxRetry=`expr $maxRetry - 1`
    sleep 10;
done

if [ "$spCount" == "$expectedSpCount" ]; then
    echo "Required number of sp(s) are present"
    testCaseStatusFlag=1
else
    echo "Required number of sp(s) are not present"
    testCaseStatusFlag=0
fi

}
check_deploy_count(){
expectedDeployCountTemp=`cat $2 | awk 'NR==2 {print}'`
expectedDeployCount=${expectedDeployCountTemp##*:}
maxRetry=10
local deployCount=-1
until [ "$deployCount" == "$expectedDeployCount" ]
do
    if [ $maxRetry == 0 ];then
        break
    fi
    deployCount=$(kubectl get deploy --no-headers -l openebs.io/storage-pool-claim=$1 -n $namespace|wc -l)
    echo -n "->"
    maxRetry=`expr $maxRetry - 1`
    sleep 10;
done

if [ "$deployCount" == "$expectedDeployCount" ]; then
    echo "Required number of pool deployment(s) are present"
    testCaseStatusFlag=1
else
    echo "Required number of pool deployment(s) are not present"
    testCaseStatusFlag=0
fi

}
check_csp_count(){
expectedCspCountTemp=`cat $2 | awk 'NR==2 {print}'`
expectedCspCount=${expectedCspCountTemp##*:}
maxRetry=10
local cspCount=-1
until [ "$cspCount" == "$expectedCspCount" ]
do
    if [ $maxRetry == 0 ];then
        break
    fi
    cspCount=$(kubectl get csp --no-headers -l openebs.io/storage-pool-claim=$1|wc -l)
    echo -n "->"
    maxRetry=`expr $maxRetry - 1`
    sleep 10;
done

if [ "$cspCount" == "$expectedCspCount" ]; then
    echo "Required number of csp(s) are present"
    testCaseStatusFlag=1
else
    echo "Required number of csp(s) are not present"
    testCaseStatusFlag=0
fi
}
if [ "$1" == "" ];then
        echo "Please give the full specified path of test folder"
        exit 1
fi

testCaseList=`cat $1preferences`
for dir in $testCaseList;do
        dirName=$1$dir
        echo $dirName
    for fileName in $dirName;do
    SPC_YAML=$dirName/`ls $fileName | grep claim`
    Expected_Output_File=$dirName/`ls $fileName | grep config`
    testCase=${dirName##*/}
    echo -e "------------------------------------------------------------"
    echo -e "${YELLOW}Running test case:"$testCase${NC}
    echo -e "------------------------------------------------------------"
    run_test $SPC_YAML $Expected_Output_File
    testFlag=$?
    resetClusterTemp=`cat $Expected_Output_File|awk 'NR==6 {print}'`
    resetCluster=${resetClusterTemp##*:}
    if [ "$resetCluster" != "false"  ];then
        echo "Clearing all pool resources"
        kubectl delete -f $SPC_YAML
    fi

    # Clear all resources
    if [ $testFlag == 1 ];then
        echo -e "${GREEN}Test Case:"$testCase "passed${NC}"
    else
        echo -e "${RED}Test Case:"$testCase "failed${NC}"
    fi
    echo -e "------------------------------------------------------------"
    echo -e "${YELLOW}Done running test case:"$testCase${NC}
    echo -e "------------------------------------------------------------"
    done
done
