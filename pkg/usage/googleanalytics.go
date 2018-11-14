package usage

import (
	"github.com/golang/glog"
	analytics "github.com/jpillora/go-ogle-analytics"
	k8sapi "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// GAclientID is the unique code of OpenEBS project in Google Analytics
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

// Send starts a goroutine to send a single event to Google Analytics
func (e *Event) Send() {
	go func() {
		v := &versionSet{}
		if err := v.getVersion(); err != nil {
			glog.Error(err)
			return
		}
		// anonymous user identifying
		// client-id - uid of default namespace
		gaClient, _ := analytics.NewClient(GAclientID)
		gaClient.ClientID(v.id).
			// OpenEBS version details
			ApplicationID("OpenEBS").
			ApplicationVersion(v.openebsVersion).
			// K8s version

			// TODO: Find k8s Environment type
			DataSource(v.nodeType).
			ApplicationName(v.k8sArch).
			ApplicationInstallerID(v.k8sVersion).
			DocumentTitle(v.id)

		event := analytics.NewEvent(e.category, e.action)
		event.Label(e.label)
		event.Value(e.value)
		if err := gaClient.Send(event); err != nil {
			glog.Error(err)
		}
		glog.V(4).Infof("Event %s:%s sent", e.category, e.action)
	}()
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
