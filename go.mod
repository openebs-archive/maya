module github.com/openebs/maya

go 1.13

require (
	cloud.google.com/go v0.46.2 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/gofuzz v1.0.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/hcl v1.0.0
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/miekg/dns v1.1.17 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/ugorji/go/codec v1.1.7
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7 // indirect
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/sys v0.0.0-20190913121621-c3b328c6e5a7 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/grpc v1.23.1
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.17.2
	sigs.k8s.io/sig-storage-lib-external-provisioner v3.1.0+incompatible

)

replace (
	k8s.io/api => k8s.io/api v0.17.3

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3

	k8s.io/apimachinery => k8s.io/apimachinery v0.17.4-beta.0

	k8s.io/apiserver => k8s.io/apiserver v0.17.3

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.3

	k8s.io/client-go => k8s.io/client-go v0.17.3

	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.3

	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.3

	k8s.io/code-generator => k8s.io/code-generator v0.17.4-beta.0

	k8s.io/component-base => k8s.io/component-base v0.17.3

	k8s.io/cri-api => k8s.io/cri-api v0.17.4-beta.0

	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.3

	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.3

	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.3

	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.3

	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.3

	k8s.io/kubectl => k8s.io/kubectl v0.17.3

	k8s.io/kubelet => k8s.io/kubelet v0.17.3

	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.3

	k8s.io/metrics => k8s.io/metrics v0.17.3

	k8s.io/node-api => k8s.io/node-api v0.17.3

	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.3

	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.3

	k8s.io/sample-controller => k8s.io/sample-controller v0.17.3
)
