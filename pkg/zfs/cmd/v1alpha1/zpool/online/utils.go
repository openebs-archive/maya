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

package ponline

// SetPool method set the Pool field of PoolOnline object.
func (p *PoolOnline) SetPool(Pool string) {
	p.Pool = Pool
}

// SetDevice method set the Device field of PoolOnline object.
func (p *PoolOnline) SetDevice(dev string) {
	p.Device = append(p.Device, dev)
}

// SetShouldExpand method set the ShouldExpand field of PoolOnline object.
func (p *PoolOnline) SetShouldExpand(ShouldExpand bool) {
	p.ShouldExpand = ShouldExpand
}

// SetCommand method set the Command field of PoolOnline object.
func (p *PoolOnline) SetCommand(Command string) {
	p.Command = Command
}

// GetPool method get the Pool field of PoolOnline object.
func (p *PoolOnline) GetPool() string {
	return p.Pool
}

// GetDevice method get the Device field of PoolOnline object.
func (p *PoolOnline) GetDevice() []string {
	return p.Device
}

// GetShouldExpand method get the ShouldExpand field of PoolOnline object.
func (p *PoolOnline) GetShouldExpand() bool {
	return p.ShouldExpand
}

// GetCommand method get the Command field of PoolOnline object.
func (p *PoolOnline) GetCommand() string {
	return p.Command
}
