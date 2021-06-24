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
)

// SendGoogleAnalytic sends a single usage metric to Google Analytics with some
// compulsory fields defined in Google Analytics API
// bindings(jpillora/go-ogle-analytics)
func (u *Usage) SendGoogleAnalytic() error {
	// Instantiate a Gclient with the tracking ID
	// Un-wrap the gaClient struct back here
	gaClient, err := analytics.NewClient(u.Gclient.trackID)
	if err != nil {
		return nil
	}
	gaClient.ClientID(u.clientID).
		CampaignSource(u.campaignSource).
		CampaignName(u.campaignName).
		CampaignContent(u.clientID).
		ApplicationID(u.appID).
		ApplicationVersion(u.appVersion).
		DataSource(u.dataSource).
		ApplicationName(u.appName).
		ApplicationInstallerID(u.appInstallerID).
		DocumentTitle(u.documentTitle)
	// Un-wrap the Event struct back here
	event := analytics.NewEvent(u.category, u.action)
	event.Label(u.label)
	event.Value(u.value)
	return gaClient.Send(event)
}
