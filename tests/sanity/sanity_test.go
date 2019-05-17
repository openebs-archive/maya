// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sanity

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	"github.com/openebs/maya/tests/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	cstorpvname string
	jivapvname  string
	present     bool
)

var _ = Describe("Sanity", func() {
	Context("should update storageclass", func() {
		It("should update storageclass wih replica count to 1", func() {
			// TODO
			// Remove this from artfacts

			// Fetching the openebs component artifacts
			artifacts, err := artifacts.GetArtifactsListUnstructuredFromFile(artifacts.SingleReplicaSC)
			Expect(err).ShouldNot(HaveOccurred())

			// Installing the artifacts to kubernetes cluster
			for _, artifact := range artifacts {
				cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
				_, err := cu.Apply(artifact)
				Expect(err).ShouldNot(HaveOccurred())
				time.Sleep(waitTime * time.Second)
			}
		})
	})

	Context("should create cstor volume", func() {
		It("should apply the cstor pvc", func() {
			// TODO
			// Remove this from artfacts

			// Getting the PVC artifact
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CStorPVCArtifacts)
			Expect(err).ShouldNot(HaveOccurred())

			// Installing the artifacts to kubernetes cluster
			cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			pvc, err := cu.Apply(artifact)
			Expect(err).ShouldNot(HaveOccurred())

			// Waiting for pvc to get bound to pv
			status := false
			for i := 0; i < 500; i++ {
				p := k8s.GetResource(k8s.GroupVersionResourceFromGVK(pvc), pvc.GetNamespace())
				pvc, err = p.Get(pvc.GetName(), metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				phase, present, err := unstructured.NestedString(pvc.Object, "status", "phase")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(present).To(BeTrue())
				if phase == "Bound" {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("pvc " + pvc.GetName() + " was not bound in expected time")
			}
			cstorpvname, present, err = unstructured.NestedString(pvc.Object, "spec", "volumeName")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(present).To(BeTrue())
			Expect(cstorpvname).ShouldNot(BeEmpty())
		})

		It("should create cstor target", func() {
			// Waiting for target to get ready
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())

				if kubernetes.CheckForPod(*pods, "target") {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("target pods is not up in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should create cstor volume replica CR", func() {
			// TODO
			// Remove this from artfacts

			// Getting the cstor volume replica artifacts
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CVRArtifact)
			Expect(err).ShouldNot(HaveOccurred())
			l := k8s.ListResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			u, err := l.List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(u.Items) != 1 {
				Fail("count of cstor volume replica should be 1")
			}
		})

		It("should create cstor replica CR", func() {
			// TODO
			// Remove this from artfacts

			// Getting the cstor replica artifacts
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CRArtifact)
			Expect(err).ShouldNot(HaveOccurred())
			l := k8s.ListResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			u, err := l.List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(u.Items) != 1 {
				Fail("count of cstor replica should be 1")
			}
		})

		It("should create cstor target service", func() {
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())

			svcs, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(svcs.Items) != 1 {
				Fail("count of cstor service should be 1")
			}
		})
	})

	Context("should delete cstor volume", func() {
		It("should delete cstor pvc", func() {
			time.Sleep(waitTime * time.Second)
			// TODO
			// Remove this from artfacts

			// Getting the PVC artifact
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CStorPVCArtifacts)
			Expect(err).ShouldNot(HaveOccurred())

			// Deleting the artifacts to kubernetes cluster
			d := k8s.DeleteResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			err = d.Delete(artifact)
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(waitTime * time.Second)
		})

		It("should delete the cstor target", func() {
			time.Sleep(waitTime * time.Second)
			// Waiting for target to get deleted
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())
				if len(pods.Items) == 0 {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("target pods is not deleted in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should delete cstor volume replica CR", func() {
			// TODO
			// Remove this from artfacts

			// Getting the cstor volume replica artifacts
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CVRArtifact)
			Expect(err).ShouldNot(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				l := k8s.ListResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
				u, err := l.List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
				Expect(err).ShouldNot(HaveOccurred())
				if len(u.Items) == 0 {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("count of cstor volume replica should be 0")
			}
		})

		It("should delete cstor replica CR", func() {
			// TODO
			// Remove this from artfacts

			// Getting the cstor replica artifacts
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.CRArtifact)
			Expect(err).ShouldNot(HaveOccurred())
			l := k8s.ListResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			u, err := l.List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(u.Items) != 0 {
				Fail("count of cstor replica should be 0")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should delete cstor target service", func() {
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())

			svcs, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/persistent-volume=" + cstorpvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(svcs.Items) != 0 {
				Fail("count of cstor service should be 0")
			}
			time.Sleep(waitTime * time.Second)
		})
	})

	Context("should create jiva volumes", func() {
		It("should apply jiva pvc", func() {
			// TODO
			// Remove this from artfacts

			// Getting the PVC artifact
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.JivaPVCArtifacts)
			Expect(err).ShouldNot(HaveOccurred())

			// Installing the artifacts to kubernetes cluster
			cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			pvc, err := cu.Apply(artifact)
			Expect(err).ShouldNot(HaveOccurred())

			// Waiting for pvc to get bound to pv
			status := false
			for i := 0; i < 500; i++ {
				p := k8s.GetResource(k8s.GroupVersionResourceFromGVK(pvc), pvc.GetNamespace())
				pvc, err = p.Get(pvc.GetName(), metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				phase, present, err := unstructured.NestedString(pvc.Object, "status", "phase")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(present).To(BeTrue())
				if phase == "Bound" {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("pvc " + pvc.GetName() + " was not bound in expected time")
			}
			jivapvname, present, err = unstructured.NestedString(pvc.Object, "spec", "volumeName")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(present).To(BeTrue())
			Expect(cstorpvname).ShouldNot(BeEmpty())
		})

		It("should create jiva controller", func() {
			// Waiting for controller to get ready
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/controller=jiva-controller,openebs.io/persistent-volume=" + jivapvname})

				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())

				if kubernetes.CheckForPod(*pods, "ctrl") {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("jiva controller pod is not up in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should create jiva replica", func() {
			// Waiting for replica to get ready
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/replica=jiva-replica,openebs.io/persistent-volume=" + jivapvname})
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())

				if kubernetes.CheckForPod(*pods, "rep") {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("jiva replica pod is not up in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should create jiva controller service", func() {
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())

			svcs, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/cas-type=jiva,openebs.io/persistent-volume=" + jivapvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(svcs.Items) != 1 {
				Fail("count of jiva service should be 1")
			}
		})
	})

	Context("should  delete jiva volume", func() {
		It("should delete jiva pvc", func() {
			// TODO
			// Remove this from artfacts

			// Getting the PVC artifact
			artifact, err := artifacts.GetArtifactUnstructuredFromFile(artifacts.JivaPVCArtifacts)
			Expect(err).ShouldNot(HaveOccurred())

			// Deleting the artifacts to kubernetes cluster
			d := k8s.DeleteResource(k8s.GroupVersionResourceFromGVK(artifact), artifact.GetNamespace())
			err = d.Delete(artifact)
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(waitTime * time.Second)
		})

		It("should delete jiva controller", func() {
			// Waiting for controller to get deleted
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/controller=jiva-controller,openebs.io/persistent-volume=" + jivapvname})
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())
				if len(pods.Items) == 0 {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("controller pod is not deleted in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should delete jiva replica", func() {
			// Waiting for controller to get deleted
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())
			status := false
			for i := 0; i < 300; i++ {
				pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/replica=jiva-replica,openebs.io/persistent-volume=" + jivapvname})
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).NotTo(BeNil())
				if len(pods.Items) == 0 {
					status = true
					break
				}
				time.Sleep(waitTime * time.Second)
			}
			if !status {
				Fail("target pod is not deleted in expected time")
			}
			time.Sleep(waitTime * time.Second)
		})

		It("should delete jiva controller service", func() {
			clientset, err := kubernetes.GetClientSet()
			Expect(err).NotTo(HaveOccurred())

			svcs, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: "openebs.io/cas-type=jiva,openebs.io/persistent-volume=" + jivapvname})
			Expect(err).ShouldNot(HaveOccurred())
			if len(svcs.Items) != 0 {
				Fail("count of jiva service should be 0")
			}
			time.Sleep(waitTime * time.Second)
		})
	})
})
