/*
Copyright 2018 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package usage

import (
	analytics "github.com/jpillora/go-ogle-analytics"
	k8sapi "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
)

const (
	// GAclientID is the unique code of OpenEBS project in Google Analytics
	GAclientID = "UA-127388617-1"
)

// Usage struct represents all information about a usage metric sent to
// Google Analytics with respect to the application
type Usage struct {
	// Embedded Event struct as we are currently only sending hits of type
	// 'event'
	Event

	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#an
	// use-case: cstor or jiva volume, or m-apiserver application
	// Embedded field for application
	Application

	// Embedded Gclient struct
	Gclient
}

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
func (u *Usage) NewEvent(c, a, l string, v int64) *Usage {
	u.category = c
	u.action = a
	u.label = l
	u.value = v
	return u
}

// Application struct holds details about the Application
type Application struct {
	// eg. project version
	appVersion string

	// eg. kubernetes version
	appInstallerID string

	// Name of the application, usage(OpenEBS/NDM)
	appID string

	// eg. usage(os-type/architecture) of system or volume's CASType
	appName string
}

// Gclient struct represents a Google Analytics hit
type Gclient struct {
	// constant tracking-id used to send a hit
	trackID string
	// anonymous client-id
	clientID string

	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#ds
	// (usecase) node-detail
	dataSource string

	// Document-title property in Google Analytics
	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#dt
	// use-case: uuid of the volume objects or a uuid to anonymously tell objects apart
	documentTitle string
}

// New returns an instance of Usage
func New() *Usage {
	return &Usage{}
}

// SetDataSource : usage(os-type, kernel)
func (u *Usage) SetDataSource(dataSource string) *Usage {
	u.dataSource = dataSource
	return u
}

// SetTrackingID Sets the GA-code for the project
func (u *Usage) SetTrackingID(track string) *Usage {
	u.trackID = track
	return u
}

// SetDocumentTitle : usecase(anonymous-id)
func (u *Usage) SetDocumentTitle(documentTitle string) *Usage {
	u.documentTitle = documentTitle
	return u
}

// SetApplicationName : usecase(os-type/arch, volume CASType)
func (u *Usage) SetApplicationName(appName string) *Usage {
	u.appName = appName
	return u
}

// SetApplicationID : usecase(OpenEBS/NDM)
func (u *Usage) SetApplicationID(appID string) *Usage {
	u.appID = appID
	return u
}

// SetApplicationVersion : usecase(project-version)
func (u *Usage) SetApplicationVersion(appVersion string) *Usage {
	u.appVersion = appVersion
	return u
}

// SetApplicationInstallerID : usecase(k8s-version)
func (u *Usage) SetApplicationInstallerID(appInstallerID string) *Usage {
	u.appInstallerID = appInstallerID
	return u
}

// SetClientID sets the anonymous user id
func (u *Usage) SetClientID(userID string) *Usage {
	u.clientID = userID
	return u
}

// Build is a builder method for Usage struct
func (u *Usage) Build() *Usage {
	// Default ApplicationID for openebs project is OpenEBS
	v := NewVersion()
	v.getVersion()
	u.SetApplicationID("OpenEBS").
		SetTrackingID(GAclientID).
		SetClientID(v.id)
	// TODO: Add condition for version over-ride
	// Case: CAS/Jiva version, etc
	return u
}

// InstallBuilder is a concrete builder for install events
func (u *Usage) InstallBuilder() *analytics.Client {
	v := NewVersion()
	clusterSize, _ := k8sapi.NumberOfNodes()
	v.getVersion()
	u.SetApplicationVersion(v.openebsVersion).
		SetApplicationName(v.k8sArch).
		SetApplicationInstallerID(v.k8sVersion).
		SetDataSource(v.nodeType).
		SetDocumentTitle(v.id).
		SetApplicationID("OpenEBS").
		NewEvent("install", "running", "nodes", int64(clusterSize))

	gaClient, _ := analytics.NewClient(GAclientID)
	// anonymous user identifying
	// client-id - uid of default namespace
	gaClient.ClientID(u.clientID).
		// OpenEBS version details
		ApplicationID(u.appID).
		ApplicationVersion(u.appVersion).
		// K8s version

		// TODO: Find k8s Environment type
		DataSource(u.dataSource).
		ApplicationName(u.appName).
		// ^ This needs to be cstor or jiva later
		ApplicationInstallerID(u.appInstallerID).
		DocumentTitle(u.documentTitle)
	// Update it to volume uuid if this is a volume-event
	// (optional) Application ID
	if u.appID != "" {
		gaClient.ApplicationID(u.appID)
	}

	//event := New()
	//event.SetCategory(u.category).
	//	SetAction(u.action).
	//	SetLabel(u.label).
	// SetValue(u.value)
	return gaClient
}
