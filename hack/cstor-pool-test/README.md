**Test script for striped dynamic cstor pool provisioning for sparse and disk type storage pool claim.**
docker.com

```test.sh``` is the script that should be used to run tests.

```striped_sparse_spc_test_cases``` folder contains the test cases for striped type of sparse claims whose provisioning mode is dynamic.

```striped_disk_spc_test_cases``` folder contains the test cases for striped type of sparse claims whose provisioning mode is dynamic.

**Steps to run the tests:**

```$./test.sh <full-specified-path-of-folder-containing-test-cases>```

The script expects anf argument which should be the full specified path of folder containing test cases.

**Note: The folder name should be appended with `/`**

Example executions of script is following:

```$./test.sh striped-disk_spc_test_cases/```

```./test.sh striped-sparse_spc_test_cases/```


