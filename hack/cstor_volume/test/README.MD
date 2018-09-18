This is a test framework for cStor Volumes. This is only a validation on how to design actual testing framework for cstor volume. It includes test cases which should be later migrated to CITF. Some test cases are not yet covered feel free to add them.

**How to use this script?**
Clone this repo and cd to this directory.
Configure operator and pool config yaml if required.
Run `bash vol-test.sh -o openebs-operator.yaml -p openebs-config.yaml`

**Important notes**
> It is assumed that there are at lease three nodes and each one of them has at least one external disk attached.
> The kubectl should be in admin context.
> The cluster is clean i.e. no existing openebs installation exists.