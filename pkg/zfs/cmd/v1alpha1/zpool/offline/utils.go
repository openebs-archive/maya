/*
Copyright 2019 The OpenEBS Authors.

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

package poffline

// SetForceOffline method set the ForceOffline field of PoolOffline object.
func (p *PoolOffline) SetForceOffline(ForceOffline bool) {
	p.ForceOffline = ForceOffline
}

// SetisTemporary method set the isTemporary field of PoolOffline object.
func (p *PoolOffline) SetisTemporary(isTemporary bool) {
	p.isTemporary = isTemporary
}

// SetPool method set the Pool field of PoolOffline object.
func (p *PoolOffline) SetPool(Pool string) {
	p.Pool = Pool
}

// SetDevice method add the device to DeviceList field of PoolOffline object.
func (p *PoolOffline) SetDevice(device string) {
	p.Devicelist = append(p.Devicelist, device)
}

// SetCommand method set the Command field of PoolOffline object.
func (p *PoolOffline) SetCommand(Command string) {
	p.Command = Command
}

// GetForceOffline method get the ForceOffline field of PoolOffline object.
func (p *PoolOffline) GetForceOffline() bool {
	return p.ForceOffline
}

// GetisTemporary method get the isTemporary field of PoolOffline object.
func (p *PoolOffline) GetisTemporary() bool {
	return p.isTemporary
}

// GetPool method get the Pool field of PoolOffline object.
func (p *PoolOffline) GetPool() string {
	return p.Pool
}

// GetDevicelist method get the Devicelist field of PoolOffline object.
func (p *PoolOffline) GetDevicelist() []string {
	return p.Devicelist
}

// GetCommand method get the Command field of PoolOffline object.
func (p *PoolOffline) GetCommand() string {
	return p.Command
}
