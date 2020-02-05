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
