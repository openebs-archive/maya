## Introduction
Backup of a cStor volume is supported through Velero plugin(http://github.com/openebs/ark-plugin).
For more information on velero, please refer https://github.com/heptio/velero.


## To create a backup
ark backup create BACKUP_NAME --snapshot-volumes

## To create a schedule backup
ark create schedule SCHEDULE_NAME --schedule="*/5 * * * *" --snapshot-volumes


## To fetch Backup status from CStor controller
```kubectl get backupcstors```

example:
```
    :~kubectl get backupcstors -n litmus
    NAME                                                         VOLUME                                     BACKUP/SCHEDULE   STATUS
    a0-pvc-a04c4885-5e87-11e9-8f95-42010a80007a                  pvc-a04c4885-5e87-11e9-8f95-42010a80007a   a0                Done
    p0-20190414152745-pvc-a04c4885-5e87-11e9-8f95-42010a80007a   pvc-a04c4885-5e87-11e9-8f95-42010a80007a   p0                Done
    p0-20190414153032-pvc-a04c4885-5e87-11e9-8f95-42010a80007a   pvc-a04c4885-5e87-11e9-8f95-42010a80007a   p0                Done
```

## To fetch information about the last successful backup
```kubectl get backupcstorlast```

example:
```
    :~kubectl get backupcstorlast -n litmus
    NAME                                          VOLUME                                     BACKUP/SCHEDULE   LASTSNAP
    a0-pvc-a04c4885-5e87-11e9-8f95-42010a80007a   pvc-a04c4885-5e87-11e9-8f95-42010a80007a   a0                a0
    p0-pvc-a04c4885-5e87-11e9-8f95-42010a80007a   pvc-a04c4885-5e87-11e9-8f95-42010a80007a   p0                p0-20190414153032
```
