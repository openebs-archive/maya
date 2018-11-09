package usage

import (
	"github.com/golang/glog"
	analytics "github.com/jpillora/go-ogle-analytics"
	k8sapi "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	openebsversion "github.com/openebs/maya/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// GAClientID is the unique code of OpenEBS project in Google Analytics
	GAclientID = "UA-127388617-1"
)

// Event is a represents usage of OpenEBS
// Event contains all the query param fields when hits is of type='event'
// Ref: https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#ec
type Event struct {
	// (Required) Event Category, ec
	category string
	// (Required) Event Action, ea
	action string
	// (Optional) Event Label, el
	label string
	// (Optional) Event vallue, ev
	// Non negative
	value int64
}

// NewEvent returns an Event struct with eventCategory, eventAction,
// eventLabel, eventValue fields
func NewEvent(c, a, l string, v int64) *Event {
	return &Event{
		category: c,
		action:   a,
		label:    l,
		value:    v,
	}
}

// Send sends a single event to Google Analytics
func (e *Event) Send() error {
	uuid, err := getUUIDbyNS("default")
	if err != nil {
		return err
	}
	gaClient, err := analytics.NewClient(GAclientID)
	if err != nil {
		return err
	}
	k8sversion, err := k8sapi.GetServerVersion()
	if err != nil {
		return err
	}
	nodeInfo, err := k8sapi.GetOSAndKernelVersion()
	if err != nil {
		return err
	}
	glog.Infof("Kubernetes version: %s", k8sversion.GitVersion)
	glog.Infof("Node type: %s", nodeInfo)
	// anonymous user identifying
	// client-id - uid of default namespace
	gaClient.ClientID(uuid).
		// OpenEBS version details
		ApplicationID("OpenEBS").
		ApplicationVersion(openebsversion.GetVersion()).
		// K8s version

		// TODO: Find k8s Environment type
		DataSource(nodeInfo).
		ApplicationName(k8sversion.Platform).
		ApplicationInstallerID(k8sversion.GitVersion).
		DocumentTitle(uuid)

	event := analytics.NewEvent(e.category, e.action)
	event.Label(e.label)
	event.Value(e.value)
	if sendSuccessErr := gaClient.Send(event); sendSuccessErr != nil {
		glog.Errorf(string(sendSuccessErr.Error()))
		return sendSuccessErr
	}
	glog.Infof("Event %s:%s fired", e.category, e.action)
	return nil
}

// getUUIDbyNS returns the metadata.object.uid of a namespace in Kubernetes
func getUUIDbyNS(namespace string) (string, error) {
	ns := k8sapi.Namespace()
	NSstruct, err := ns.Get(namespace, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if NSstruct != nil {
		return string(NSstruct.GetObjectMeta().GetUID()), nil
	}
	return "", nil

}
