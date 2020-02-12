package main

import (
	"fmt"

	exec "github.com/openebs/maya/pkg/upgrade/executor_new"
)

func main() {
	err := exec.Exec(
		"1.6.0", "1.7.0", "cstorpoolcluster",
		"sparse-pool-1", "openebs", "openebs/", "ci",
	)
	fmt.Println(err)
}

// upgrade cstor-cspc --from-version=1.6.0 --to-version=1.7.0 --to-version-image-prefix="openebs/" --to-version-image-tag=ci sparse-pool-1 sparse-pool-2
