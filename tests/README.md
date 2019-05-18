Behavior Driven Development(BDD)

Maya makes use of ginkgo & gomega libraries to implement its integration tests.

To run the BDD test:
1) Copy kubeconfig file into path ~/.kube/config or set the KUBECONFIG env.
2) Change the current directory path to test related directory
Example:
To trigger the statefulesets test
Step1: Change the pwd to test directory.
       `cd github.com/openebs/maya/tests/sts/`
Step2: Execute the command `ginkgo -v -- -kubeconfig=/path/to/kubeconfig`

Output:
Sample example output
```
Running Suite: StatefulSet
==========================
Random Seed: 1555919486
Will run 1 of 2 specs

StatefulSet test statefulset application on cstor
  should distribute the cstor volume replicas across pools
  /home/sai/gocode/src/github.com/openebs/maya/tests/sts/sts_test.go:227
---------------------------------------------------------

Ran 1 of 1 Specs in 767.022 seconds
FAIL! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped

Ginkgo ran 1 suite in 12m49.573774959s
Test Suite Failed
```
Note: Above is the sample output how it looks when you ran `ginkgo -v -- -kubeconfig=/path/to/kubeconfig`
