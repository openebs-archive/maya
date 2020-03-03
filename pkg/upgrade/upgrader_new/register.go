/*
Copyright 2019 The OpenEBS Authors

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

package upgrader

func (u *Upgrade) registerUpgrade(kind string, obj UpgradeOptions) *Upgrade {
	u.UpgradeMap[kind] = obj
	return u
}

// RegisterAll ...
func (u *Upgrade) RegisterAll() *Upgrade {
	u.registerUpgrade("cstorpoolinstance", RegisterCstorPoolInstance)
	u.registerUpgrade("cstorpoolcluster", RegisterCstorPoolCluster)
	// u.registerUpgrade("cstorVolume", RegisterCstorVolume)
	// u.registerUpgrade("jivaVolume", RegisterJivaVolume)
	return u
}

// RegisterCstorPoolInstance ....
func RegisterCstorPoolInstance(r *ResourcePatch, c *Client) Upgrader {
	obj := NewCSPIPatch(
		WithCSPIResorcePatch(r),
		WithCSPIClient(c),
	)
	return obj
}

// RegisterCstorPoolCluster ...
func RegisterCstorPoolCluster(r *ResourcePatch, c *Client) Upgrader {
	obj := NewCSPCPatch(
		WithCSPCResorcePatch(r),
		WithCSPCClient(c),
	)
	return obj
}
