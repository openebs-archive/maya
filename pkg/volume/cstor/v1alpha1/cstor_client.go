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

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	volume "github.com/openebs/maya/pkg/client/volume/cstor/v1alpha1"
)

// client is used to invoke cstor resize client call
type client struct {
	ip     string
	volume *apis.CASVolume
}

type clientBuilder struct {
	client *client
}

// ClientBuilder forms the populates the client structure
func ClientBuilder() *clientBuilder {
	return &clientBuilder{client: &client{}}
}

func (c *clientBuilder) WithIP(ip string) *clientBuilder {
	c.client.ip = ip
	return c
}

func (c *clientBuilder) WithVolume(v *apis.CASVolume) *clientBuilder {
	c.client.volume = v
	return c
}

// Build return a pointer to client
func (c *clientBuilder) Build() *client {
	return c.client
}

// Resize will trigger a grpc call to cstor-volume-mgmt side car to resize
// volume
func (c *client) Resize() (*apis.CASVolume, error) {
	_, err := volume.ResizeVolume(c.ip, c.volume.Name, c.volume.Spec.Capacity)
	if err != nil {
		return nil, err
	}

	// we are returning the same struct that we received as input.
	// This would be modified when server replies back with some property of
	// resize volume
	return c.volume, nil
}
