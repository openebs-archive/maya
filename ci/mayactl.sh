echo "*************** Running mayactl pool list *******************************"
${MAYACTL} pool list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "*************** Running mayactl pool describe *******************************"
${MAYACTL} pool describe --poolname $POOLNAME
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"

echo "*************** Running mayactl volume list *******************************"
${MAYACTL} volume list
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"


echo "************** Running Jiva mayactl volume describe **************************"
${MAYACTL} volume describe --volname $JIVAVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

printf "\n\n"
sleep 5

echo "************** Running Jiva mayactl volume stats *************************"
${MAYACTL} volume stats --volname  $JIVAVOL -n openebs
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

echo "************** Running Cstor mayactl volume stats *************************"
sleep 30
${MAYACTL} volume stats --volname  $CSTORVOL -n openebs
rc=$?;
echo $rc
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi

#sleep 60
#echo "************** Running Jiva mayactl snapshot create **********************"
#${MAYACTL} snapshot create --volname $JIVAVOL --snapname snap1
#rc=$?;
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#printf "\n\n"
#sleep 30
#
#${MAYACTL} snapshot create --volname $JIVAVOL --snapname snap2
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#sleep 30
#
#echo "************** Running Jiva mayactl snapshot list ************************"
#${MAYACTL} snapshot list --volname $JIVAVOL
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#
#printf "\n\n"
#sleep 30
#echo "************** Running Jiva mayactl snapshot revert **********************"
#${MAYACTL} snapshot revert --volname $JIVAVOL --snapname snap1
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi
#printf "\n\n"
#sleep 10
#
#echo "************** Running Jiva mayactl snapshot list after revert ************"
#${MAYACTL} snapshot list --volname $JIVAVOL
#if [[ $rc != 0 ]]; then
#	kubectl logs --tail=10 $MAPIPOD -n openebs
#	exit $rc;
#fi


printf "\n\n"
echo "************** Running Cstor mayactl volume describe *************************"
${MAYACTL} volume describe --volname $CSTORVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=10 $MAPIPOD -n openebs
	exit $rc;
fi


echo "************** Running Jiva mayactl volume delete ************************"
${MAYACTL} volume delete --volname $JIVAVOL
rc=$?;
if [[ $rc != 0 ]]; then
	kubectl logs --tail=100 $MAPIPOD -n openebs
	exit $rc;
fi
sleep 30

printf "\n\n"
echo "************** Check if jiva replica data is cleared *************************"
if [ -f /var/openebs/$JIVAVOL/volume.meta ]; then
	#Check if the job is in progress.
	printf "\n"
	ls -lR /var/openebs
	printf "\n"
	kubectl get jobs
	printf "\n"
	kubectl get pods
	printf "\n"
else
   echo "Jiva replica data is cleared successfully"
fi

