package artifacts

import (
	"io/ioutil"
	"strings"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openebs/maya/pkg/artifact/v1alpha1"
)

// parseK8sYaml parses the kubernetes yaml and returns the objects in a UnstructuredList
func parseK8sYaml(filename string) (k8s.UnstructedList, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return k8s.UnstructedList{}, err
	}
	fileAsString := string(fileBytes[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	artifacts := v1alpha1.ArtifactList{}
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}
		f = strings.TrimSpace(f)
		artifacts.Items = append(artifacts.Items, &v1alpha1.Artifact{Doc: f})
	}
	ulist, _ := artifacts.ToUnstructuredList()
	return ulist, err
}

// GetArtifactsUnstructured returns the unstructured list of openebs components
func GetArtifactsUnstructured() ([]*unstructured.Unstructured, error) {
	ulist, err := parseK8sYaml("../artifacts/openebs-ci.yaml")
	if err != nil {
		return nil, err
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items, err
}
