#!bin/bash

source ../util.sh

cleanUp()
{
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

# sleep as image pulling takes time
sleep 10
pvStatus=
try=1
printf "%s" "Checking status of pvc:"
until [ "$pvStatus" == "Pending" ] || [ "$try" == "5" ] || [ "$pvStatus" == "Bound" ]; do
    printf " %s" "$try"
    pvStatus=$(kubectl get pvc openebs-pvc -o jsonpath='{.status.phase}')
    try=`expr $try + 1`
    sleep 5
done

echo ""

echo PVC status: $pvStatus
if [ "$pvStatus" == "Bound" ]; then
    echo Unexpected: pvc in running state
    exit 1
fi

if [ "$pvStatus" == "Pending" ]; then
   error=$(kubectl describe pvc openebs-pvc | grep "not enough pool" | wc -l)
   echo Grepping \'not enough pool\' resulted $error value/s
   if [ "$error" == "0" ]; then
       echo Expected error not found
       cleanUp
       exit 1
   fi
fi
cleanUp
exit 0
