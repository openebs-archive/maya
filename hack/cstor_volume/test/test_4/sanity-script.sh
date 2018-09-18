#!bin/bash
# remove if the file already exists
rm write-status.txt
#run the command
kubectl exec $appName -- touch /var/lib/openebsvol/text.file 
#write the exit code to file
echo $? > write-status.txt
