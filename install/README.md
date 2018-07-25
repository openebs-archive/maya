## Install
This document consists of series of steps to install openebs components. 

## Step 1 - Install OpenEBS Operator
`kubectl apply -f openebs-operator.yaml`

## Step 2 - Get Disk Info
`kubectl get disks`

## Step 3 - Create StoragePoolClaim
```yaml
apiVersion: openebs.io/v1alpha1
kind: StoragePoolClaim
metadata:
  name: gkepooldemo
  annotations:
    openebs.io/create-template: cast-standard-cstorpool-0.6.0
spec:
  name: gkepooldemo
  type: openebs-cstor
  maxPools: 3
  poolSpec:
    poolType: striped
    cacheFile: /tmp/gkepooldemo.cache
    overProvisioning: false
  disks:
    diskList:
      # list of disks obtained from kubectl get disks
      - disk-4268137899842721d2d4fc0c16c3b138
      - disk-49c3f6bfe9906e8db04adda12815375c
      - disk-552af787ba458a22fa0cf355d17da885
      - disk-99cde73d1defa35375029e8164e974e0
      - disk-9b8d2c6eba0d15a9434f37d49dba4076
      - disk-e49a2876b475a73a314f6eeeeb6d5c53
```

## Step 4 - Create StorageClass
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-standard
  annotations:
    cas.openebs.io/create-volume-template: cast-standard-cstor-create-0.7.0
    cas.openebs.io/config: |
      - name: StoragePoolClaim
        value: "gkepooldemo"
provisioner: openebs.io/provisioner-iscsi
```

## Step 5 - Deploy mysql application
`kubectl apply -f percona-deploy.yaml`
