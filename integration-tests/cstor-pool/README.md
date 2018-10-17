<ToDO> Add Readme Doc to run integration test suite
All cstor pool integration tests reside in this directory
A sample spc yaml ( Used in test case #1)
---
apiVersion: openebs.io/v1alpha1
kind: StoragePoolClaim
metadata:
  name: sparse-claim-auto
spec:
  name: sparse-claim-auto
  type: sparse
  maxPools: 3
  poolSpec:
    poolType: striped