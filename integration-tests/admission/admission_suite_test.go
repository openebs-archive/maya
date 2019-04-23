package admission_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admission Suite")
}
