package command_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/openebs/maya/command"
	"os"
)

var _ = Describe("GetVolAnnotations", func() {
	var server *ghttp.Server
	var returnedVolume Volume
	var statusCode int
	var annotations *Annotations

	BeforeEach(func() {
		server = ghttp.NewServer()
		os.Setenv("MAPI_ADDR", "http://"+server.Addr())

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/latest/volumes/info/VOLUME"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, &returnedVolume),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("When the server returns a volume", func() {
		BeforeEach(func() {
			annotations = &Annotations{
				Iqn:              "iqn.2016-09.com.openebs.jiva:vol",
				TargetPortal:     "10.99.73.74:3260",
				VolSize:          "1G",
				ClusterIP:        "10.99.73.74",
				ReplicaCount:     "2",
				ControllerStatus: "Running",
				ReplicaStatus:    "",
				ControllerIP:     "",
				VolAddr:          "",
				Replicas:         "",
			}
			returnedVolume = Volume{
				Metadata: struct {
					Annotations       interface{} `json:"annotations"`
					CreationTimestamp interface{} `json:"creationTimestamp"`
					Name              string      `json:"name"`
				}{
					Annotations: annotations,
				},
			}
			statusCode = 200
		})

		It("returns the annotations associated with the volume", func() {
			value, err := GetVolAnnotations("VOLUME")
			Expect(err).To(BeNil())
			Expect(value).Should(Equal(annotations))
		})
	})

	Context("when the server returns 500", func() {
		BeforeEach(func() {
			statusCode = 500
		})

		It("throws errors", func() {
			value, err := GetVolAnnotations("VOLUME")
			Expect(value).Should(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the server returns 503", func() {
		BeforeEach(func() {
			statusCode = 503
		})

		It("throws errors", func() {
			value, err := GetVolAnnotations("VOLUME")
			Expect(value).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
