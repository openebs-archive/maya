package artifacts

import (
	"errors"
	"io/ioutil"
	"strings"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openebs/maya/pkg/artifact/v1alpha1"
)

// ArtifactSource holds the path to fetch artifacts
type ArtifactSource string

const (
	OpenEBSArtifacts  ArtifactSource = "../artifacts/openebs-ci.yaml"
	CStorPVCArtifacts ArtifactSource = "../artifacts/cstor-pvc.yaml"
	JivaPVCArtifacts  ArtifactSource = "../artifacts/jiva-pvc.yaml"
	SingleReplicaSC   ArtifactSource = "../artifacts/single-replica.yaml"
	CVRArtifact       ArtifactSource = "../artifacts/cvr-schema.yaml"
	CRArtifact        ArtifactSource = "../artifacts/cr-schema.yaml"
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

// GetArtifactsListUnstructured returns the unstructured list of openebs components
func GetArtifactsListUnstructured(a ArtifactSource) ([]*unstructured.Unstructured, error) {
	ulist, err := parseK8sYaml(string(a))
	if err != nil {
		return nil, err
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items, err
}

// GetArtifactUnstructured returns the unstructured list of openebs components
func GetArtifactUnstructured(a ArtifactSource) (*unstructured.Unstructured, error) {
	ulist, err := parseK8sYaml(string(a))
	if err != nil {
		return nil, err
	}
	if len(ulist.Items) != 1 {
		return nil, errors.New("more than one artifacts found")
	}
	nList := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{})
	return nList.Items[0], nil
}
